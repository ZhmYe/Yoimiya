package plugin

import (
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	"sync"
)

const INVALID_TID int = -1

// BParam 用来记录Buffer里的每一个item
type BParam struct {
	tID     int                     // task的ID，用于将结果返回给task
	witness witness.Witness         // fullWitness,不需要assignment
	extra   []constraint.ExtraValue // todo 这里是否需要？是否可以集成到上面的witness
}

// Buffer Buffer里面可以存assignment，但还是要对应具体的task
type Buffer struct {
	sID   int
	items []BParam
	mutex sync.Mutex
}

func NewBuffer(id int) *Buffer {
	return &Buffer{
		sID:   id,
		items: make([]BParam, 0),
	}
}
func (b *Buffer) IsEmpty() bool {
	return len(b.items) == 0
}
func (b *Buffer) Push(tID int, witness witness.Witness, extra []constraint.ExtraValue) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.items = append(b.items, BParam{tID: tID, witness: witness, extra: extra})
}

// Pop 去除第一个buffer元素
// todo 加锁？
func (b *Buffer) Pop() BParam {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if b.IsEmpty() {
		panic("Buffer is empty!!!")
	}
	e := b.items[0]
	if len(b.items) == 1 {
		b.items = make([]BParam, 0)
	} else {
		b.items = b.items[1:]
	}
	return e
}
