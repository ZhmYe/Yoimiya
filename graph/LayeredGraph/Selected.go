package LayeredGraph

// SelectedQueue 备选集，根据节点的深度得到优先级队列
// 每个队列内部根据出度进行排序
type SelectedQueue struct {
	depth int     // 该队列对应的深度
	nodes []*Node // 队列中的所有节点
}

func NewSelectedQueue(depth int) *SelectedQueue {
	return &SelectedQueue{
		depth: depth,
		nodes: make([]*Node, 0),
	}
}

// func (q *SelectedQueue)

type SelectedSet struct {
}
