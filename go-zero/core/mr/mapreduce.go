package mr

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/lang"
)

const (
	defaultWorkers = 16
	minWorkers     = 1
)

var (
	// ErrCancelWithNil is an error that mapreduce was cancelled with nil.
	ErrCancelWithNil = errors.New("mapreduce cancelled with nil")
	// ErrReduceNoOutput is an error that reduce did not output a value.
	ErrReduceNoOutput = errors.New("reduce not writing value")
)

type (
	// ForEachFunc is used to do element processing, but no output.
	ForEachFunc func(item interface{})
	// GenerateFunc is used to let callers send elements into source.
	GenerateFunc func(source chan<- interface{})
	// MapFunc is used to do element processing and write the output to writer.
	MapFunc func(item interface{}, writer Writer)
	// MapperFunc is used to do element processing and write the output to writer,
	// use cancel func to cancel the processing.
	MapperFunc func(item interface{}, writer Writer, cancel func(error))
	// ReducerFunc is used to reduce all the mapping output and write to writer,
	// use cancel func to cancel the processing.
	ReducerFunc func(pipe <-chan interface{}, writer Writer, cancel func(error))
	// VoidReducerFunc is used to reduce all the mapping output, but no output.
	// Use cancel func to cancel the processing.
	VoidReducerFunc func(pipe <-chan interface{}, cancel func(error))
	// Option defines the method to customize the mapreduce.
	Option func(opts *mapReduceOptions)

	mapperContext struct {
		ctx       context.Context
		mapper    MapFunc
		source    <-chan interface{}
		panicChan *onceChan
		collector chan<- interface{}
		doneChan  <-chan lang.PlaceholderType
		workers   int
	}

	mapReduceOptions struct {
		ctx     context.Context
		workers int
	}

	// Writer interface wraps Write method.
	Writer interface {
		Write(v interface{})
	}
)

// Finish runs fns parallelly, cancelled on any error.
func Finish(fns ...func() error) error {
	if len(fns) == 0 {
		return nil
	}

	return MapReduceVoid(func(source chan<- interface{}) {
		for _, fn := range fns {
			source <- fn
		}
	}, func(item interface{}, writer Writer, cancel func(error)) {
		fn := item.(func() error)
		if err := fn(); err != nil {
			cancel(err)
		}
	}, func(pipe <-chan interface{}, cancel func(error)) {
	}, WithWorkers(len(fns)))
}

// FinishVoid runs fns parallelly.
func FinishVoid(fns ...func()) {
	if len(fns) == 0 {
		return
	}

	ForEach(func(source chan<- interface{}) {
		for _, fn := range fns {
			source <- fn
		}
	}, func(item interface{}) {
		fn := item.(func())
		fn()
	}, WithWorkers(len(fns)))
}

// ForEach maps all elements from given generate but no output.
func ForEach(generate GenerateFunc, mapper ForEachFunc, opts ...Option) {
	options := buildOptions(opts...)
	panicChan := &onceChan{channel: make(chan interface{})}
	source := buildSource(generate, panicChan)
	collector := make(chan interface{}, options.workers)
	done := make(chan lang.PlaceholderType)

	go executeMappers(mapperContext{
		ctx: options.ctx,
		mapper: func(item interface{}, writer Writer) {
			mapper(item)
		},
		source:    source,
		panicChan: panicChan,
		collector: collector,
		doneChan:  done,
		workers:   options.workers,
	})

	for {
		select {
		case v := <-panicChan.channel:
			panic(v)
		case _, ok := <-collector:
			if !ok {
				return
			}
		}
	}
}

// MapReduce maps all elements generated from given generate func,
// and reduces the output elements with given reducer.
// MapReduce 工具主要用来对批量数据进行并发的处理
// 以此来提升服务的性能
func MapReduce(generate GenerateFunc, mapper MapperFunc, reducer ReducerFunc,
	opts ...Option) (interface{}, error) {
	panicChan := &onceChan{channel: make(chan interface{})}
	source := buildSource(generate, panicChan)
	return mapReduceWithPanicChan(source, panicChan, mapper, reducer, opts...)
}

// MapReduceChan maps all elements from source, and reduce the output elements with given reducer.
func MapReduceChan(source <-chan interface{}, mapper MapperFunc, reducer ReducerFunc,
	opts ...Option) (interface{}, error) {
	panicChan := &onceChan{channel: make(chan interface{})}
	return mapReduceWithPanicChan(source, panicChan, mapper, reducer, opts...)
}

// MapReduceChan maps all elements from source, and reduce the output elements with given reducer.
func mapReduceWithPanicChan(source <-chan interface{}, panicChan *onceChan, mapper MapperFunc,
	reducer ReducerFunc, opts ...Option) (interface{}, error) {
	options := buildOptions(opts...)
	// output is used to write the final result
	output := make(chan interface{})
	defer func() {
		// reducer can only write once, if more, panic
		for range output {
			panic("more than one element written in reducer")
		}
	}()

	// collector is used to collect data from mapper, and consume in reducer
	collector := make(chan interface{}, options.workers)
	// if done is closed, all mappers and reducer should stop processing
	done := make(chan lang.PlaceholderType)
	writer := newGuardedWriter(options.ctx, output, done)
	var closeOnce sync.Once
	// use atomic.Value to avoid data race
	var retErr errorx.AtomicError
	finish := func() {
		closeOnce.Do(func() {
			close(done)
			close(output)
		})
	}
	// cancel方法，mapper和reducer中都可以调用该方法
	// 调用后主线程收到close信号会立马返回
	cancel := once(func(err error) {
		if err != nil {
			retErr.Set(err)
		} else {
			retErr.Set(ErrCancelWithNil)
		}

		drain(source)
		// 调用close(ouput)主线程收到Done信号，立马返回
		finish()
	})

	// reducer 对 goroutine 对应 mapper写入collector的数据进行处理
	// 如果reducer中没有手动调用writer.Write
	// 则最终会执行finish方法对output进行close避免死锁
	go func() {
		defer func() {
			drain(collector)
			if r := recover(); r != nil {
				panicChan.write(r)
			}
			finish()
		}()

		reducer(collector, writer, cancel)
	}()

	// 消费buildSource产生的数据，每一个item都会起一个goroutine单独处理
	// 默认最大并发数为16
	// 可以通过 WithWorkers 进行设置
	go executeMappers(mapperContext{
		ctx: options.ctx,
		mapper: func(item interface{}, w Writer) {
			mapper(item, w, cancel)
		},
		source:    source,
		panicChan: panicChan,
		collector: collector,
		doneChan:  done,
		workers:   options.workers,
	})

	select {
	case <-options.ctx.Done():
		cancel(context.DeadlineExceeded)
		return nil, context.DeadlineExceeded
	case v := <-panicChan.channel:
		panic(v)
	case v, ok := <-output:
		if err := retErr.Load(); err != nil {
			return nil, err
		} else if ok {
			return v, nil
		} else {
			return nil, ErrReduceNoOutput
		}
	}
}

// MapReduceVoid maps all elements generated from given generate,
// and reduce the output elements with given reducer.
func MapReduceVoid(generate GenerateFunc, mapper MapperFunc, reducer VoidReducerFunc, opts ...Option) error {
	_, err := MapReduce(generate, mapper, func(input <-chan interface{}, writer Writer, cancel func(error)) {
		reducer(input, cancel)
	}, opts...)
	if errors.Is(err, ErrReduceNoOutput) {
		return nil
	}

	return err
}

// WithContext customizes a mapreduce processing accepts a given ctx.
func WithContext(ctx context.Context) Option {
	return func(opts *mapReduceOptions) {
		opts.ctx = ctx
	}
}

// WithWorkers customizes a mapreduce processing with given workers.
func WithWorkers(workers int) Option {
	return func(opts *mapReduceOptions) {
		if workers < minWorkers {
			opts.workers = minWorkers
		} else {
			opts.workers = workers
		}
	}
}

func buildOptions(opts ...Option) *mapReduceOptions {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func buildSource(generate GenerateFunc, panicChan *onceChan) chan interface{} {
	// 通过buildSource方法通过执行generate(参数为无缓冲channel)产生数据
	// 并返回无缓冲的channel，mapper会从该channel中读取数据
	source := make(chan interface{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChan.write(r)
			}
			close(source)
		}()

		generate(source)
	}()

	return source
}

// drain drains the channel.
func drain(channel <-chan interface{}) {
	// drain the channel
	for range channel {
	}
}

func executeMappers(mCtx mapperContext) {
	var wg sync.WaitGroup
	defer func() {
		wg.Wait() // 保证所有的item都处理完成
		close(mCtx.collector)
		drain(mCtx.source)
	}()

	var failed int32
	pool := make(chan lang.PlaceholderType, mCtx.workers)
	// 将mapper处理完的数据写入collector
	writer := newGuardedWriter(mCtx.ctx, mCtx.collector, mCtx.doneChan)
	for atomic.LoadInt32(&failed) == 0 {
		select {
		case <-mCtx.ctx.Done(): // 当调用了cancel会触发立即返回
			return
		case <-mCtx.doneChan:
			return
		case pool <- lang.Placeholder: // 控制最大并发数
			item, ok := <-mCtx.source
			if !ok {
				<-pool
				return
			}

			wg.Add(1)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						atomic.AddInt32(&failed, 1)
						mCtx.panicChan.write(r)
					}
					wg.Done()
					<-pool
				}()
				// 对item进行处理
				// 处理完调用writer.Write把结果写入collector对应的channel中
				mCtx.mapper(item, writer)
			}()
		}
	}
}

func newOptions() *mapReduceOptions {
	return &mapReduceOptions{
		ctx:     context.Background(),
		workers: defaultWorkers,
	}
}

func once(fn func(error)) func(error) {
	once := new(sync.Once)
	return func(err error) {
		once.Do(func() {
			fn(err)
		})
	}
}

type guardedWriter struct {
	ctx     context.Context
	channel chan<- interface{}
	done    <-chan lang.PlaceholderType
}

func newGuardedWriter(ctx context.Context, channel chan<- interface{},
	done <-chan lang.PlaceholderType) guardedWriter {
	return guardedWriter{
		ctx:     ctx,
		channel: channel,
		done:    done,
	}
}

func (gw guardedWriter) Write(v interface{}) {
	select {
	case <-gw.ctx.Done():
		return
	case <-gw.done:
		return
	default:
		gw.channel <- v
	}
}

type onceChan struct {
	channel chan interface{}
	wrote   int32
}

func (oc *onceChan) write(val interface{}) {
	if atomic.AddInt32(&oc.wrote, 1) > 1 {
		return
	}

	oc.channel <- val
}
