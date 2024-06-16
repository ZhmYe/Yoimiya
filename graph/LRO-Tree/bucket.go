package LRO_Tree

// Bucket 用于选择节点，一定程度后将已选中的节点弹出
type Bucket struct {
	threshold int // 阈值
	split     int // split计数
	count     int // 当前split的数量
	//items     []*LroNode
}

func NewBucket() *Bucket {
	return &Bucket{
		threshold: -1,
		split:     0,
	}
}
func (b *Bucket) SetThreshold(t int) {
	b.threshold = t
}
func (b *Bucket) Check(node *LroNode) bool {
	// todo 这里的逻辑
	// 首先判断是否已经满了
	// 如果满了
	tolerant := 1 // 这里暂定最大容忍度为1
	if b.count >= b.threshold+tolerant {
		return false
	}
	return true
}
func (b *Bucket) Add(node *LroNode) {
	if !b.Check(node) {
		//b.items = append(b.items, node)
		// 赋予节点split
		//node.SetSplit(b.split)
		b.Pop()
	}
	node.SetSplit(b.split)
	b.count++
}

// Pop 弹出所有的items
func (b *Bucket) Pop() {
	b.split++
	b.count = 0
	//for _, node := range b.items {
	//	node.SetSplit(b.split)
	//}
	//b.items = make([]*LroNode, 0)
}