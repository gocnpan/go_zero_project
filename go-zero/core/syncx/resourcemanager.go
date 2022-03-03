package syncx

import (
	"io"
	"sync"

	"github.com/zeromicro/go-zero/core/errorx"
)

// A ResourceManager is a manager that used to manage resources.
type ResourceManager struct {
	resources    map[string]io.Closer
	// 让多个同一 key 的操作，针对同一内容时，获取首个操作的值进行共享
	// SingleFlight lets the concurrent calls with the same key to share the call result.
	// For example, A called F, before it's done, B called F. Then B would not execute F,
	// and shared the result returned by F which called by A.
	// The calls with the same key are dependent, concurrent calls share the returned values.
	// A ------->calls F with key<------------------->returns val
	// B --------------------->calls F with key------>returns val
	singleFlight SingleFlight
	lock         sync.RWMutex
}

// NewResourceManager returns a ResourceManager.
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:    make(map[string]io.Closer),
		singleFlight: NewSingleFlight(),
	}
}

// Close closes the manager.
// Don't use the ResourceManager after Close() called.
func (manager *ResourceManager) Close() error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	var be errorx.BatchError
	for _, resource := range manager.resources {
		if err := resource.Close(); err != nil {
			be.Add(err)
		}
	}

	// release resources to avoid using it later
	manager.resources = nil

	return be.Err()
}

// GetResource returns the resource associated with given key.
func (manager *ResourceManager) GetResource(key string, create func() (io.Closer, error)) (io.Closer, error) {
	// 这里保证，针对同一数据库连接，只创建一次，后续针对同一数据库连接的创建，均共享首次
	// 通过关闭释放资源
	val, err := manager.singleFlight.Do(key, func() (interface{}, error) {
		manager.lock.RLock()
		resource, ok := manager.resources[key]
		manager.lock.RUnlock()
		if ok {
			return resource, nil
		}

		resource, err := create()
		if err != nil {
			return nil, err
		}

		manager.lock.Lock()
		manager.resources[key] = resource
		manager.lock.Unlock()

		return resource, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(io.Closer), nil
}
