package LRO_Tree

// Bucket 用于选择节点，一定程度后将已选中的节点弹出
type Bucket struct {
	threshold int // 阈值
	split     int // split计数
	items     []*LroNode
}

func NewBucket() *Bucket {
	return &Bucket{
		threshold: -1,
		split:     -1,
	}
}
func (b *Bucket) SetThreshold(t int) {
	b.threshold = t
}

// Pop 弹出所有的items,将这些节点的split赋予
func (b *Bucket) Pop() {
	b.split++
	for _, node := range b.items {
		node.SetSplit(b.split)
	}
	b.items = make([]*LroNode, 0)
}
