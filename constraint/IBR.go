package constraint

type Item struct {
	CallData  []uint32  // 这里直接存calldata不需要instruction
	BluePrint Blueprint // 这里直接把instruction和对应的Item放在一起
}

// IBR Instruction-BluePrint Record 代替DataRecord
type IBR struct {
	items []Item // 记录所有的Instruction+BluePrint，这里包括了Calldata
	//CoeffTable     cs_bn254.CoeffTable
	//CommitmentInfo constraint.Commitments
	// 这俩部分是公用的，直接单独拿出来
}

func NewIBR() *IBR {
	return &IBR{items: make([]Item, 0)}
}
func (r *IBR) Items() []Item {
	return r.items
}
func (r *IBR) Append(data []uint32, b Blueprint) {
	item := Item{
		CallData:  data,
		BluePrint: b,
	}
	r.items = append(r.items, item)
}

//func (r *IBR) GetCoeffTable() cs_bn254.CoeffTable {
//	return r.CoeffTable
//}
//func (r *IBR) GetCommitmentInfo() constraint.Commitments {
//	return r.CommitmentInfo
//}
