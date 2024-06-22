package constraint

type Item struct {
	CallData        []uint32  // 这里直接存calldata不需要instruction
	BluePrint       Blueprint // 这里直接把instruction和对应的Item放在一起
	isForwardOutput bool      // 当前instruction是否需要被传递给后面的电路
}

func (i *Item) IsForwardOutput() bool {
	return i.isForwardOutput
}

// IBR Instruction-BluePrint Record 代替DataRecord
type IBR struct {
	items   []Item // 记录所有的Instruction+BluePrint，这里包括了Calldata
	witness []int  // todo 这里的witness目前仅是input
	//CoeffTable     cs_bn254.CoeffTable
	//CommitmentInfo constraint.Commitments
	// 这俩部分是公用的，直接单独拿出来
}

func NewIBR() *IBR {
	return &IBR{items: make([]Item, 0), witness: make([]int, 0)}
}
func (r *IBR) Items() []Item {
	return r.items
}
func (r *IBR) Append(data []uint32, b Blueprint, isForwardOutput bool) {
	item := Item{
		CallData:        data,
		BluePrint:       b,
		isForwardOutput: isForwardOutput,
	}
	r.items = append(r.items, item)
}
func (r *IBR) SetWitness(w []int) {
	r.witness = w
}
func (r *IBR) GetWitness() []int {
	return r.witness
}

//func (r *IBR) GetCoeffTable() cs_bn254.CoeffTable {
//	return r.CoeffTable
//}
//func (r *IBR) GetCommitmentInfo() constraint.Commitments {
//	return r.CommitmentInfo
//}
