package LRO_Tree

const FAKE_ROOT_ID int = -2
const SPLIT_UNSET int = -1

// LroNodeState 节点状态
type LroNodeState = int

const (
	NotVisited LroNodeState = iota
	Normal
	Middle
)

// LroNode 每个instruction都是L·R=O的结果
// 记录节点的L和R是哪些节点
// 简单来说也就是节点的父节点有哪些，因为L和R是等价的
// 但L和R之间可能是有连接关系的，因此把L和R全部统计在一个地方
// 用Node的depth来说明他们的先后遍历顺序
type LroNode struct {
	LR      []*LroNode   // 按照depth排序
	depth   int          // O的深度
	degree  int          // O的出度
	O       int          // iID
	visited LroNodeState // 是否被访问过
	root    bool         // 是否为root
	split   int          // O被放到哪个split
}

func NewLroNode(iID int) *LroNode {
	return &LroNode{
		LR:      make([]*LroNode, 0),
		depth:   -1,
		degree:  0,
		O:       iID,
		visited: NotVisited,
		root:    true,
		split:   SPLIT_UNSET,
	}
}
func (n *LroNode) NotRoot() {
	n.root = false
}
func (n *LroNode) IsRoot() bool {
	return n.root
}

// SetLR 设置父节点
func (n *LroNode) SetLR(previousNodes []*LroNode) {
	if len(previousNodes) == 0 {
		n.depth = 0
	}
	n.LR = previousNodes // 这里先不排序
	maxLevel := -1
	for _, node := range n.LR {
		if node.depth > maxLevel {
			maxLevel = node.depth
		}
	}
	maxLevel++
	n.depth = maxLevel
}
func (n *LroNode) SignToMiddle() {
	if !n.IsVisited() {
		panic("Node should be Visited!!!")
	}
	n.visited = Middle
}
func (n *LroNode) IsMiddle() bool {
	return n.visited == Middle
}
func (n *LroNode) Visit() {
	if n.IsVisited() {
		panic("This Node has been Visited!!!")
	}
	n.visited = Normal
}
func (n *LroNode) TryVisit() bool {
	if !n.IsVisited() {
		//n.Visit()
		// 可以访问,n.Visit()在后序遍历时候调用
		return true
	}
	//} else if !n.IsMiddle() {
	//	n.SignToMiddle()
	//	return false
	//}
	return false
}
func (n *LroNode) IsVisited() bool {
	return n.visited != NotVisited
}
func (n *LroNode) AddDegree() {
	n.degree++
}
func (n *LroNode) SetSplit(s int) {
	if n.split != SPLIT_UNSET {
		panic("This Node's Split has been Set!!!")
	}
	n.split = s
}

// Ergodic 后续遍历
// todo 节点遍历逻辑，后序遍历
func (n *LroNode) Ergodic(b *Bucket) {
	// 1. 先遍历depth小的父节点,depth一样先遍历出度小的节点
	// 2. 每次在添加O后判断是否弹出bucket:
	//    (1) 如果还没满，添加O
	//    (2) 如果满了，留一个（这个待定）位置，判断下一个O是否也已经遍历完LR，如果不是则添加O并弹出Bucket，反之直接弹出Bucket

	// LR里的所有节点按照depth和出度排序
	sort := func(a *LroNode, b *LroNode) bool {

		if a.depth < b.depth {
			return true
		} else if a.depth == b.depth {
			return a.degree <= b.degree
		}
		return false
	}
	se := NewSortEngine(sort)
	n.LR = se.Sort(n.LR).([]*LroNode)
	// 后续遍历
	for _, node := range n.LR {
		if !node.TryVisit() {
			continue
		}
		node.Ergodic(b)
	}
	//fmt.Println(n.O)
	n.Visit()
	b.Add(n)
	for _, node := range n.LR {
		node.CheckMiddle(n.split)
	}
}
func (n *LroNode) CheckMiddle(split int) {
	if n.split != split {
		n.visited = Middle
	}
}

//func (n *LroNode) CheckFinish() bool {
//	for _, node := range n.LR {
//		if !node.Visit() {
//			return false
//		}
//	}
//}