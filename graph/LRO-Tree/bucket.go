package LRO_Tree

// Bucket 用于选择节点，一定程度后将已选中的节点弹出
type Bucket struct {
	threshold     int // 阈值
	split         int // split计数
	count         int // 当前split的数量
	nbInstruction int
	total         int
	LEVEL         []int // 这里就不建一个level了数量比较少
	//items     []*LroNode
}

func NewBucket() *Bucket {
	return &Bucket{
		threshold:     -1,
		split:         0,
		LEVEL:         []int{0},
		nbInstruction: 0,
		total:         0,
	}
}
func (b *Bucket) SetTotalNbWires(total int) {
	b.total = total
}
func (b *Bucket) SetThreshold(cut int) {
	nbWires := b.total
	b.threshold = RoundUpSplit(nbWires, cut)
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

// UpdateSplitLevel todo 这里可以加上Split之间的关系

func (b *Bucket) UpdateSplitLevel(pre int, next int) {
	preLevel := b.LEVEL[pre]
	//fmt.Println(pre, next)
	if len(b.LEVEL) == next {
		b.LEVEL = append(b.LEVEL, preLevel+1)
	} else {
		if preLevel > b.LEVEL[next] {
			b.LEVEL[next] = preLevel
		}
	}
}
func (b *Bucket) Add(node *InstructionNode) {
	if node.O == FAKE_ROOT_ID {
		return
	}
	if !b.Check(node) {
		//b.items = append(b.items, node)
		// 赋予节点split
		//node.SetSplit(b.split)
		b.Pop()
	}
	node.SetSplit(b.split)
	//fmt.Println(node.O, node.split)
	b.nbInstruction++
	b.count++
}
func (b *Bucket) Alloc(node *InputNode) {
	//if NotIn := !(node.SetSplit(b.split)); NotIn {
	//	b.count++
	//}
	node.SetSplit(b.split)
	//fmt.Println(node.wireID, node.splits)

}

// Pop 弹出所有的items
func (b *Bucket) Pop() {
	// todo 如果最后一个电路太小，就合并到前面去
	if b.nbInstruction*10 < b.total*9 {
		b.split++
		b.LEVEL = append(b.LEVEL, b.LEVEL[len(b.LEVEL)-1])
	}
	b.count = 0
}
func (b *Bucket) CheckLastSplitIsEmpty() int {
	if b.count == 0 {
		b.split--
	}
	return b.split + 1
}

// GetSplitLevel 获取Split之间的并行关系
func (b *Bucket) GetSplitLevel() [][]int {
	levels := make([][]int, 0)
	for s, level := range b.LEVEL {
		for level >= len(levels) {
			levels = append(levels, make([]int, 0))
		}
		levels[level] = append(levels[level], s)
	}
	return levels
}
