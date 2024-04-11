package MisalignedParalleling

type Buffer struct {
	id    int   // Buffer id，对应slot id
	queue []int // taskID
	//mutex sync.Mutex                                   // queue的同步
}

func NewBuffer(id int) *Buffer {
	return &Buffer{
		id:    id,
		queue: make([]int, 0),
	}
}
func (b *Buffer) IsEmpty() bool {
	return len(b.queue) == 0
}
func (b *Buffer) Push(id int) {
	b.queue = append(b.queue, id)
}

// Pop 去除第一个buffer元素
func (b *Buffer) Pop() int {
	if b.IsEmpty() {
		return -1
	}
	e := b.queue[0]
	if len(b.queue) == 1 {
		b.queue = make([]int, 0)
	} else {
		b.queue = b.queue[1:]
	}
	return e
}
