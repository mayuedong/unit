package unit

import "sync"

type SyncQueue[T comparable] struct {
	running bool
	max     int
	queue   *List[T]
	mutex   sync.Mutex
	empty   *sync.Cond
	full    *sync.Cond
}

func (r *SyncQueue[T]) Init(max int) {
	r.max = max
	r.empty = sync.NewCond(&r.mutex)
	r.full = sync.NewCond(&r.mutex)
	r.queue = &List[T]{max: max}
	r.running = true
}

func (r *SyncQueue[T]) Add(val T) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for r.queue.Full() && r.running {
		r.full.Wait()
	}

	if r.running {
		r.queue.Add(val)
		r.empty.Signal()
	}
}
func (r *SyncQueue[T]) AddMore(vals []T) {
	l := &List[T]{}
	for _, val := range vals {
		l.Push(val)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	for r.queue.Full() && r.running {
		r.full.Wait()
	}

	if r.running {
		r.queue = r.queue.AddList(l)
		r.empty.Signal()
	}
}
func (r *SyncQueue[T]) Push(val T) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for r.queue.Full() && r.running {
		r.full.Wait()
	}

	if r.running {
		r.queue.Push(val)
		r.empty.Signal()
	}
}
func (r *SyncQueue[T]) PushMore(vals []T) {
	l := &List[T]{}
	for _, val := range vals {
		l.Push(val)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	for r.queue.Full() && r.running {
		r.full.Wait()
	}

	if r.running {
		r.queue = r.queue.PushList(l)
		r.empty.Signal()
	}
}

func (r *SyncQueue[T]) Pop() (val T, ok bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for r.queue.Empty() && r.running {
		r.empty.Wait()
	}

	val, ok = r.queue.PopHead()
	if r.max > 0 {
		r.full.Signal()
	}
	return
}

func (r *SyncQueue[T]) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.running = false
	r.empty.Broadcast()
	if r.max > 0 {
		r.full.Broadcast()
	}
}

func (r *SyncQueue[T]) Count() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.queue.Count()
}
