package unit

type QueueNode[K comparable, V any] struct {
	k K
	v V
}

func (r *QueueNode[K, V]) getKey() (key K) {
	return r.k
}

func (r *QueueNode[K, V]) getVal() (val V) {
	return r.v
}

type FastQueue[K comparable, V any] struct {
	l List[*QueueNode[K, V]]
	m map[K]*ListNode[*QueueNode[K, V]]
}

func (r *FastQueue[K, V]) Init(max int) {
	r.l.Init(max)
	r.m = make(map[K]*ListNode[*QueueNode[K, V]])
}

func (r *FastQueue[K, V]) Count() int {
	return len(r.m)
}

func (r *FastQueue[K, V]) Push(k K, v V) {
	r.m[k] = r.l.Push(&QueueNode[K, V]{k, v})
}

func (r *FastQueue[K, V]) Add(k K, v V) {
	r.m[k] = r.l.Add(&QueueNode[K, V]{k, v})
}

func (r *FastQueue[K, V]) Del(k K) {
	if node, ok := r.m[k]; ok {
		r.l.del(node)
		delete(r.m, k)
	}
}

func (r *FastQueue[K, V]) GetHead() (val V, ok bool) {
	if node, ok := r.l.GetHead(); ok {
		return node.getVal(), true
	}
	return
}

func (r *FastQueue[K, V]) GetTail() (val V, ok bool) {
	if node, ok := r.l.GetTail(); ok {
		return node.getVal(), true
	}
	return
}

func (r *FastQueue[K, V]) PopHead() (val V, ok bool) {
	if node, ok := r.l.PopHead(); ok {
		delete(r.m, node.getKey())
		return node.getVal(), true
	}
	return
}

func (r *FastQueue[K, V]) PopTail() (val V, ok bool) {
	if node, ok := r.l.PopTail(); ok {
		delete(r.m, node.getKey())
		return node.getVal(), true
	}
	return
}

func (r *FastQueue[K, V]) GetHeadRotateTail() (val V, ok bool) {
	if node, ok := r.l.GetHead(); ok {
		r.l.RotateHeadToTail()
		return node.getVal(), true
	}
	return
}

func (r *FastQueue[K, V]) GetTailRotateHead() (val V, ok bool) {
	if node, ok := r.l.GetTail(); ok {
		r.l.RotateTailToHead()
		return node.getVal(), true
	}
	return
}
