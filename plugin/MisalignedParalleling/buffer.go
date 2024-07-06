package MisalignedParalleling

import (
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
)

// BParam 用来记录Buffer里的每一个item
type BParam struct {
	tID     int                     // task的ID，用于将结果返回给task
	witness witness.Witness         // fullWitness,不需要assignment
	extra   []constraint.ExtraValue // todo 这里是否需要？是否可以集成到上面的witness
}

// Buffer Buffer里面可以存assignment，但还是要对应具体的task
type Buffer struct {
	sID   int
	items chan BParam
}

func NewBuffer(id int, capacity int) *Buffer {
	return &Buffer{
		sID:   id,
		items: make(chan BParam, capacity),
	}
}

//func (b *Buffer) IsEmpty() bool {
//	return len(b.items) == 0
//}

func (b *Buffer) Push(tID int, witness witness.Witness, extra []constraint.ExtraValue) {
	b.items <- BParam{tID: tID, witness: witness, extra: extra}
}

func (b *Buffer) Pop() BParam {
	//if b.IsEmpty() {
	//	panic("Buffer is empty!!!")
	//}
	return <-b.items
}
