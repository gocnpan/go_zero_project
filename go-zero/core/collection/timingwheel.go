package collection

import (
	"container/list"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/lang"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/core/timex"
)

const drainWorkers = 8

type (
	// Execute defines the method to execute the task.
	Execute func(key, value interface{})

	// A TimingWheel is a timing wheel object to schedule tasks.
	// timingWheel 靠定时器推动
	// 时间前进的同时会取出当前时间格中 list「双向链表」的task，传递到 execute 中执行。
	// 因为是是靠 internal 固定时间刻度推进，可能就会出现：
	// 一个 60s 的task，internal = 1s，这样就会空跑59次loop。
	//
	// 而在扩展时间上，采取 circle 分层，这样就可以不断复用原有的 numSlots
	// 因为定时器在不断 loop，而执行可以把上层的 slot 下降到下层
	// 在不断 loop 中就可以执行到上层的task。
	// 这样的设计可以在不创造额外的数据结构，突破长时间的限制。
	// 参考：https://www.jianshu.com/p/ba6056069f9f
	// collection 中的 timingWheel ，维护一个存放任务组的数组，每一个槽都维护一个存储task的双向链表。
	// 开始执行时，计时器每隔指定时间执行一个槽里面的tasks。
	TimingWheel struct {
		interval      time.Duration			// 单个时间格时间间隔
		ticker        timex.Ticker				// 定时器，做时间推动，以interval为单位推进
		slots         []*list.List				// 时间轮
		timers        *SafeMap				// 存储task{key, value}的map [执行execute所需要的参数]
		tickedPos     int						// at previous virtual circle
		numSlots      int						// 初始化 slots num
		execute       Execute					// 执行函数
		// 以下几个channel是做task传递的
		setChannel    chan timingEntry
		moveChannel   chan baseEntry
		removeChannel chan interface{}
		drainChannel  chan func(key, value interface{})
		stopChannel   chan lang.PlaceholderType
	}

	timingEntry struct {
		baseEntry
		value   interface{}
		circle  int
		diff    int
		removed bool
	}

	baseEntry struct {
		delay time.Duration
		key   interface{}
	}

	positionEntry struct {
		pos  int
		item *timingEntry
	}

	timingTask struct {
		key   interface{}
		value interface{}
	}
)

// NewTimingWheel returns a TimingWheel.
func NewTimingWheel(interval time.Duration, numSlots int, execute Execute) (*TimingWheel, error) {
	if interval <= 0 || numSlots <= 0 || execute == nil {
		return nil, fmt.Errorf("interval: %v, slots: %d, execute: %p", interval, numSlots, execute)
	}

	return newTimingWheelWithClock(interval, numSlots, execute, timex.NewTicker(interval))
}

func newTimingWheelWithClock(interval time.Duration, numSlots int, execute Execute, ticker timex.Ticker) (
	*TimingWheel, error) {
	tw := &TimingWheel{
		interval:      interval,
		ticker:        ticker,
		slots:         make([]*list.List, numSlots),
		timers:        NewSafeMap(),
		tickedPos:     numSlots - 1,
		execute:       execute,
		numSlots:      numSlots,
		setChannel:    make(chan timingEntry),
		moveChannel:   make(chan baseEntry),
		removeChannel: make(chan interface{}),
		drainChannel:  make(chan func(key, value interface{})),
		stopChannel:   make(chan lang.PlaceholderType),
	}

	// 把 slot 中存储的 list 全部准备好
	tw.initSlots()
	// 开启异步协程，使用 channel 来做task通信和传递
	go tw.run()

	return tw, nil
}

// Drain drains all items and executes them.
func (tw *TimingWheel) Drain(fn func(key, value interface{})) {
	tw.drainChannel <- fn
}

// MoveTimer moves the task with the given key to the given delay.
func (tw *TimingWheel) MoveTimer(key interface{}, delay time.Duration) {
	if delay <= 0 || key == nil {
		return
	}

	tw.moveChannel <- baseEntry{
		delay: delay,
		key:   key,
	}
}

// RemoveTimer removes the task with the given key.
func (tw *TimingWheel) RemoveTimer(key interface{}) {
	if key == nil {
		return
	}

	tw.removeChannel <- key
}

// SetTimer sets the task value with the given key to the delay.
func (tw *TimingWheel) SetTimer(key, value interface{}, delay time.Duration) {
	if delay <= 0 || key == nil {
		return
	}

	tw.setChannel <- timingEntry{
		baseEntry: baseEntry{
			delay: delay,
			key:   key,
		},
		value: value,
	}
}

// Stop stops tw.
func (tw *TimingWheel) Stop() {
	close(tw.stopChannel)
}

func (tw *TimingWheel) drainAll(fn func(key, value interface{})) {
	runner := threading.NewTaskRunner(drainWorkers)
	for _, slot := range tw.slots {
		for e := slot.Front(); e != nil; {
			task := e.Value.(*timingEntry)
			next := e.Next()
			slot.Remove(e)
			e = next
			if !task.removed {
				runner.Schedule(func() {
					fn(task.key, task.value)
				})
			}
		}
	}
}

func (tw *TimingWheel) getPositionAndCircle(d time.Duration) (pos, circle int) {
	steps := int(d / tw.interval)
	pos = (tw.tickedPos + steps) % tw.numSlots
	circle = (steps - 1) / tw.numSlots

	return
}

func (tw *TimingWheel) initSlots() {
	for i := 0; i < tw.numSlots; i++ {
		tw.slots[i] = list.New()
	}
}

// delay < internal：因为 < 单个时间精度，表示这个任务已经过期，需要马上执行
// 针对改变的 delay：
//   new >= old：<newPos, newCircle, diff>
//   newCircle > 0：计算diff，并将 circle 转换为 下一层，故diff + numslots
//   如果只是单纯延迟时间缩短，则将老的task标记删除，重新加入list，等待下一轮loop被execute
func (tw *TimingWheel) moveTask(task baseEntry) {
	// timers: Map => 通过key获取 [positionEntry「pos, task」]
	val, ok := tw.timers.Get(task.key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	// {delay < interval} => 延迟时间比一个时间格间隔还小
	// 没有更小的刻度
	// 说明任务应该立即执行
	if task.delay < tw.interval {
		threading.GoSafe(func() {
			tw.execute(timer.item.key, timer.item.value)
		})
		return
	}

	// 如果 > interval
	// 则通过延迟时间 delay
	// 计算其出时间轮中的 new pos, circle
	pos, circle := tw.getPositionAndCircle(task.delay)
	if pos >= timer.pos {
		timer.item.circle = circle
		// 记录前后的移动offset
		// 为了后面过程重新入队
		timer.item.diff = pos - timer.pos
	} else if circle > 0 {
		// 转移到下一层，将 circle 转换为 diff 一部分
		circle--
		timer.item.circle = circle
		// 因为是一个数组，要加上 numSlots [也就是相当于要走到下一层]
		timer.item.diff = tw.numSlots + pos - timer.pos
	} else {
		// 如果 offset 提前了，此时 task 也还在第一层
		// 标记删除老的 task，并重新入队，等待被执行
		timer.item.removed = true
		newItem := &timingEntry{
			baseEntry: task,
			value:     timer.item.value,
		}
		tw.slots[pos].PushBack(newItem)
		tw.setTimerPosition(pos, newItem)
	}
}

// 定时器 「每隔 internal 会执行一次」
func (tw *TimingWheel) onTick() {
	// 每次执行更新一下当前执行 tick 位置
	tw.tickedPos = (tw.tickedPos + 1) % tw.numSlots
	// 获取此时 tick位置 中的存储task的双向链表
	l := tw.slots[tw.tickedPos]
	tw.scanAndRunTasks(l)
}

func (tw *TimingWheel) removeTask(key interface{}) {
	val, ok := tw.timers.Get(key)
	if !ok {
		return
	}

	timer := val.(*positionEntry)
	timer.item.removed = true
	tw.timers.Del(key)
}

func (tw *TimingWheel) run() {
	for {
		select {
		// 定时器做时间推动 -> scanAndRunTasks()
		case <-tw.ticker.Chan():
			tw.onTick()
		// add task 会往 setChannel 输入task
		case task := <-tw.setChannel:
			tw.setTask(&task)
		case key := <-tw.removeChannel:
			tw.removeTask(key)
		case task := <-tw.moveChannel:
			tw.moveTask(task)
		case fn := <-tw.drainChannel:
			tw.drainAll(fn)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimingWheel) runTasks(tasks []timingTask) {
	if len(tasks) == 0 {
		return
	}

	go func() {
		for i := range tasks {
			threading.RunSafe(func() {
				tw.execute(tasks[i].key, tasks[i].value)
			})
		}
	}()
}

func (tw *TimingWheel) scanAndRunTasks(l *list.List) {
	// 存储目前需要执行的task{key, value}  [execute所需要的参数，依次传递给execute执行]
	var tasks []timingTask

	for e := l.Front(); e != nil; {
		task := e.Value.(*timingEntry)
		// 标记删除，在 scan 中做真正的删除 「删除map的data」
		if task.removed {
			next := e.Next()
			l.Remove(e)
			e = next
			continue
		} else if task.circle > 0 {
			// 当前执行点已经过期
			// 但是同时不在第一层，所以当前层即然已经完成了，就会降到下一层
			// 但是并没有修改 pos
			task.circle--
			e = e.Next()
			continue
		} else if task.diff > 0 {
			// 因为之前已经标注了diff，需要再进入队列
			next := e.Next()
			l.Remove(e)
			// (tw.tickedPos+task.diff)%tw.numSlots
			// cannot be the same value of tw.tickedPos
			pos := (tw.tickedPos + task.diff) % tw.numSlots
			tw.slots[pos].PushBack(task)
			tw.setTimerPosition(pos, task)
			task.diff = 0
			e = next
			continue
		}

		// 以上的情况都是不能执行的情况，能够执行的会被加入tasks中
		tasks = append(tasks, timingTask{
			key:   task.key,
			value: task.value,
		})
		next := e.Next()
		l.Remove(e)
		tw.timers.Del(task.key)
		e = next
	}

	// for range tasks，然后把每个 task->execute 执行即可
	tw.runTasks(tasks)
}

func (tw *TimingWheel) setTask(task *timingEntry) {
	if task.delay < tw.interval {
		task.delay = tw.interval
	}

	if val, ok := tw.timers.Get(task.key); ok {
		entry := val.(*positionEntry)
		entry.item.value = task.value
		tw.moveTask(task.baseEntry)
	} else {
		pos, circle := tw.getPositionAndCircle(task.delay)
		task.circle = circle
		tw.slots[pos].PushBack(task)
		tw.setTimerPosition(pos, task)
	}
}

func (tw *TimingWheel) setTimerPosition(pos int, task *timingEntry) {
	if val, ok := tw.timers.Get(task.key); ok {
		timer := val.(*positionEntry)
		timer.item = task
		timer.pos = pos
	} else {
		tw.timers.Set(task.key, &positionEntry{
			pos:  pos,
			item: task,
		})
	}
}
