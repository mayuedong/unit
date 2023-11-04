package unit

import (
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	stopped   chan bool
	running   atomic.Bool
	threads   int
	onceClose sync.Once
	queue     SyncQueue[anyEvent]
}

type anyEvent interface {
	Callback()
}

func NewWorkerPool(queueSize, threads int) *WorkerPool {
	r := &WorkerPool{threads: threads, queue: SyncQueue[anyEvent]{}, stopped: make(chan bool)}
	r.queue.Init(queueSize)
	r.running.Store(true)
	for i := 0; i < r.threads; i++ {
		go func(i int) {
			for {
				if task, ok := r.take(); ok {
					task.Callback()
				} else if !r.running.Load() {
					r.stopped <- true
					return
				}
			}
		}(i)
	}
	return r
}

func (r *WorkerPool) Close() {
	r.onceClose.Do(func() {
		r.running.Store(false)
		r.queue.Stop()
		for i := 0; i < r.threads; i++ {
			<-r.stopped
		}

	})
}

func (r *WorkerPool) Add(task anyEvent) {
	if r.running.Load() {
		r.queue.Add(task)
	}
}

func (r *WorkerPool) Push(task anyEvent) {
	if r.running.Load() {
		r.queue.Push(task)
	}

}

func (r *WorkerPool) take() (anyEvent, bool) {
	return r.queue.Pop()
}

type WorkerPoolGroup struct {
	index     atomic.Uint32
	length    uint32
	onceClose sync.Once
	pools     []*WorkerPool
}

func NewWorkerPoolGroup(count, queueSize, threads int) *WorkerPoolGroup {
	r := &WorkerPoolGroup{length: uint32(count)}
	for i := 0; i < count; i++ {
		r.pools = append(r.pools, NewWorkerPool(queueSize, threads))
	}
	return r
}
func (r *WorkerPoolGroup) Close() {
	r.onceClose.Do(func() {
		for _, pool := range r.pools {
			pool.Close()
		}
	})
}
func (r *WorkerPoolGroup) Push(task anyEvent) {
	r.pools[r.index.Add(1)%r.length].Push(task)
}

func (r *WorkerPoolGroup) Add(task anyEvent) {
	r.pools[r.index.Add(1)%r.length].Add(task)
}
