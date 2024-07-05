package LRO_Tree

type InputNode struct {
	wireID  int
	degree  int
	splits  []int
	visited bool
}

func NewInputNode(wireID int) *InputNode {
	return &InputNode{
		wireID:  wireID,
		degree:  0,
		splits:  make([]int, 0),
		visited: false,
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
func (n *InputNode) SetSplit(s int) bool {
	for _, split := range n.splits {
		if split == s {
			return true
		}
	}
	n.splits = append(n.splits, s)
	return false
	//_, exist := n.splits[s]
	//if exist {
	//	return true
	//}
	//n.splits[s] = true
	//return false
}
func (n *InputNode) TryVisit() bool {
	return true
}
func (n *InputNode) Visited() {
	n.visited = true
}
func (n *InputNode) CheckMiddle(s int, b *Bucket) {
	n.SetSplit(s)
}
func (n *InputNode) IsMiddle() bool {
	return false
}
func (n *InputNode) Ergodic(b *Bucket) {
	b.Alloc(n)
	return
}
func (n *InputNode) GetSplits() []int {
	//splits := make([]int, 0)
	//for split, _ := range n.splits {
	//	splits = append(splits, split)
	//}
	return n.splits
}