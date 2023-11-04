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

func (r *List[T]) Init(max int) {
	r.max = max
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
	if 2 > r.Count() {
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
	if 2 > r.Count() {
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
