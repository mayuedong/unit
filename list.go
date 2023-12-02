package unit

type ListNode[T comparable] struct {
	data T
	prev *ListNode[T]
	next *ListNode[T]
}

func (r *ListNode[T]) GetVal() (val T) {
	return r.data
}

type List[T comparable] struct {
	head  *ListNode[T]
	tail  *ListNode[T]
	count int
	max   int
}

func (r *List[T]) Count() int {
	return r.count
}

func (r *List[T]) Empty() bool {
	return 0 == r.count
}

func (r *List[T]) Full() bool {
	if r.max > 0 && r.count >= r.max {
		return true
	}
	return false
}

func (r *List[T]) Add(val T) (node *ListNode[T]) {
	node = &ListNode[T]{data: val}
	if 0 == r.count {
		r.head = node
		r.tail = node
	} else {
		node.next = r.head
		r.head.prev = node
		r.head = node
	}
	r.count++
	return
}

func (r *List[T]) PopHead() (val T, ok bool) {
	if 0 == r.count {
		return
	}
	node := r.head
	r.del(node)
	return node.data, true
}

func (r *List[T]) GetHead() (val T, ok bool) {
	if 0 == r.count {
		return
	}
	return r.head.data, true
}

func (r *List[T]) Push(val T) (node *ListNode[T]) {
	node = &ListNode[T]{data: val}
	if 0 == r.count {
		r.head = node
		r.tail = node
	} else {
		node.prev = r.tail
		r.tail.next = node
		r.tail = node
	}
	r.count++
	return
}

func (r *List[T]) PopTail() (val T, ok bool) {
	if 0 == r.count {
		return
	}
	node := r.tail
	r.del(node)
	return node.data, true
}

func (r *List[T]) GetTail() (val T, ok bool) {
	if 0 == r.count {
		return
	}
	return r.tail.data, true
}

func (r *List[T]) RotateHeadToTail() {
	if 2 > r.count {
		return
	}
	head := r.head
	r.head = head.next
	r.head.prev = nil
	r.tail.next = head
	head.next = nil
	head.prev = r.tail
	r.tail = head
}

func (r *List[T]) RotateTailToHead() {
	if 2 > r.count {
		return
	}
	tail := r.tail
	r.tail = tail.prev
	r.tail.next = nil
	r.head.prev = tail
	tail.prev = nil
	tail.next = r.head
	r.head = tail

}

func (r *List[T]) del(node *ListNode[T]) {
	if nil != node.prev {
		node.prev.next = node.next
	} else {
		r.head = node.next
	}
	if nil != node.next {
		node.next.prev = node.prev
	} else {
		r.tail = node.prev
	}
	r.count--
}

func (r *List[T]) FromHeadFindNode(op func(T) bool) (node *ListNode[T]) {
	for node = r.head; node != nil; node = node.next {
		if op(node.data) {
			return node
		}
	}
	return
}

func (r *List[T]) FromTailFindNode(op func(T) bool) (node *ListNode[T]) {
	for node = r.tail; node != nil; node = node.prev {
		if op(node.data) {
			return node
		}
	}
	return
}

func (r *List[T]) InsertPrev(val T, oldNode *ListNode[T]) {
	r.insert(val, oldNode, false)
}

func (r *List[T]) InsertNext(val T, oldNode *ListNode[T]) {
	r.insert(val, oldNode, true)
}

func (r *List[T]) insert(val T, oldNode *ListNode[T], after bool) {
	node := &ListNode[T]{data: val}
	if after {
		node.prev = oldNode
		node.next = oldNode.next
		if r.tail == oldNode {
			r.tail = node
		}
	} else {
		node.next = oldNode
		node.prev = oldNode.prev
		if r.head == oldNode {
			r.head = node
		}
	}
	if node.prev != nil {
		node.prev.next = node
	}
	if node.next != nil {
		node.next.prev = node
	}
	r.count++
}
func (r *List[T]) PushList(o *List[T]) *List[T] {
	r.insertList(o)
	return r
}
func (r *List[T]) AddList(o *List[T]) *List[T] {
	if nil == o || o.Count() <= 0 {
		return r
	}
	o.max = r.max
	o.insertList(r)
	return o
}
func (r *List[T]) insertList(o *List[T]) {
	if o.Count() <= 0 {
		return
	}

	o.head.prev = r.tail

	if nil != r.tail {
		r.tail.next = o.head
	} else {
		r.head = o.head
	}
	r.tail = o.tail
	r.count += o.Count()

	o.head = nil
	o.tail = nil
	o.count = 0
}
