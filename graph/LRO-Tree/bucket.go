package LRO_Tree

// Bucket 用于选择节点，一定程度后将已选中的节点弹出
type Bucket struct {
	threshold     int // 阈值
	split         int // split计数
	count         int // 当前split的数量
	nbInstruction int
	total         int
	//items     []*LroNode
}

func NewBucket() *Bucket {
	return &Bucket{
		threshold:     -1,
		split:         0,
		nbInstruction: 0,
		total:         0,
	}
}
func (b *Bucket) SetTotalNbInstruction(total int) {
	b.total = total
}
func (b *Bucket) SetThreshold(t int) {
	b.threshold = t
}
func (b *Bucket) Check(node *InstructionNode) bool {
	// todo 这里的逻辑
	// 首先判断是否已经满了
	// 如果满了
	if node.O == FAKE_ROOT_ID {
		return false
	}
	tolerant := 1 // 这里暂定最大容忍度为1
	if b.count >= b.threshold+tolerant {
		return false
	}
	return true
}
func (b *Bucket) Add(node *InstructionNode) {
	if !b.Check(node) {
		//b.items = append(b.items, node)
		// 赋予节点split
		//node.SetSplit(b.split)
		b.Pop()
	}
	node.SetSplit(b.split)
	b.nbInstruction++
	b.count++
}
func (b *Bucket) Alloc(node *InputNode) {
	if NotIn := !(node.SetSplit(b.split)); NotIn {
		b.count++
	}

}

// Pop 弹出所有的items
func (b *Bucket) Pop() {
	// todo 如果最后一个电路太小，就合并到前面去
	if b.nbInstruction*10 < b.total*9 {
		b.split++
	}
	b.count = 0
}
func (b *Bucket) CheckLastSplitIsEmpty() {
	if b.count == 0 {
		b.split--
	}
}
