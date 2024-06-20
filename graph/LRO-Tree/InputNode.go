package LRO_Tree

type InputNode struct {
	wireID int
	degree int
	splits map[int]bool
}

func NewInputNode(wireID int) *InputNode {
	return &InputNode{
		wireID: wireID,
		degree: 0,
	}
}
func (n *InputNode) Depth() int {
	return 0
}
func (n *InputNode) Degree() int {
	return n.degree
}
func (n *InputNode) AddDegree() {
	n.degree++
}
func (n *InputNode) IsRoot() bool {
	return false
}
func (n *InputNode) NotRoot() {
	return
}
func (n *InputNode) SetSplit(s int) {
	n.splits[s] = true
}
func (n *InputNode) TryVisit() bool {
	return true
}
func (n *InputNode) CheckMiddle(s int) {
	return
}
func (n *InputNode) IsMiddle() bool {
	return len(n.splits) == 0
}
func (n *InputNode) Ergodic(b *Bucket) {
	return
}
func (n *InputNode) GetSplits() []int {
	splits := make([]int, 0)
	for split, _ := range n.splits {
		splits = append(splits, split)
	}
	return splits
}
