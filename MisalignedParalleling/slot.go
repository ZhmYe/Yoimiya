package MisalignedParalleling

// todo 目前的实现还只能支持cut=2
// todo 这里的逻辑

// Slot 用来表示一个电路不同部分的槽位，当slot被占用时不能被其它的task使用
type Slot struct {
	cs       PackedConstraintSystem // 该槽位的电路
	isFilled bool                   // 是否被占用
	id       int                    // 槽位id，暂时用来debug

}

func NewSlot(id int, cs PackedConstraintSystem) *Slot {
	return &Slot{
		id:       id,
		isFilled: false,
		cs:       cs,
	}
}

// CheckFilled 判断是否被占用
func (s *Slot) CheckFilled() bool {
	return s.isFilled
}
func (s *Slot) GetConstraintSystem() PackedConstraintSystem {
	return s.cs
}
