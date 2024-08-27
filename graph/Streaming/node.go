package Streaming

// ConstraintNode 图中的每一个节点，代表一个Instruction或者Input
type ConstraintNode struct {
	wireID  int               // 对应的wire
	degree  int               // 度数
	depth   int               // 深度
	visited bool              // 用于ordering
	child   []*ConstraintNode // 子节点
	parent  []int             // 父节点id，在Instruction插入时已经知晓，无需额外计算
}

func NewConstraintNode(id int, previousIDs []int) *ConstraintNode {
	return &ConstraintNode{
		wireID:  id,
		degree:  0,
		depth:   -1,
		visited: false,
		child:   make([]*ConstraintNode, 0),
		parent:  previousIDs,
	}
}
func (n *ConstraintNode) AddDegree() {
	n.degree++
}
func (n *ConstraintNode) InsertChild(node *ConstraintNode) {
	n.child = append(n.child, node)
}
func (n *ConstraintNode) Degree() int {
	return n.degree
}
func (n *ConstraintNode) Depth() int {
	return n.depth
}
func (n *ConstraintNode) UpdateDepth(parentDepth int) {
	if n.depth < parentDepth+1 {
		n.depth = parentDepth + 1
	}
}
