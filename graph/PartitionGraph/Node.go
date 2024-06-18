package PartitionGraph

const SignMiddle int = -2

// Node 图上的节点，表示一个instruction，记录其所有“父节点”
// 图中每个level存放这些Node
type Node struct {
	iID int // instruction id
	//piIDs []int // 所有父节点的id，在图上是子节点
	//child  []*Node // 所有前置instruction，在图上是子节点
	level     int // level记录在这里，避免外面需要使用Map
	partition int // 记录Node在哪个split里
	//isRoot    bool // 记录是否为root
	// todo 如果判断是否是middle
}

func NewNode(iID int) Node {
	return Node{
		iID: iID,
		//child:  make([]*Node, 0),
		level:     -1,
		partition: -1,
		//isRoot:    true,
		//deepest: -1,
	}
}

//	func (n *Node) IsRoot() bool {
//		return n.isRoot
//	}
//
//	func (n *Node) NotRoot() {
//		n.isRoot = false
//	}
func (n *Node) AssignPartitionAndLevel(partition int, level int) {
	n.partition = partition
	n.level = level
}
