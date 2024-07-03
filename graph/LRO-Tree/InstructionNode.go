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

// InstructionNode 每个instruction都是L·R=O的结果
// 记录节点的L和R是哪些节点
// 简单来说也就是节点的父节点有哪些，因为L和R是等价的
// 但L和R之间可能是有连接关系的，因此把L和R全部统计在一个地方
// 用Node的depth来说明他们的先后遍历顺序

type InstructionNode struct {
	LR      []LroNode    // 按照depth排序
	depth   int          // O的深度
	degree  int          // O的出度
	O       int          // iID
	visited LroNodeState // 是否被访问过
	root    bool         // 是否为root
	split   int          // O被放到哪个split
}

func NewInstructionNode(iID int) *InstructionNode {
	return &InstructionNode{
		LR:      make([]LroNode, 0),
		depth:   -1,
		degree:  0,
		O:       iID,
		visited: NotVisited,
		root:    true,
		split:   SPLIT_UNSET,
	}
}
func (n *InstructionNode) Depth() int {
	return n.depth
}
func (n *InstructionNode) Degree() int {
	return n.degree
}
func (n *InstructionNode) NotRoot() {
	n.root = false
}
func (n *InstructionNode) IsRoot() bool {
	return n.root
}

// SetLR 设置父节点
func (n *InstructionNode) SetLR(previousNodes []LroNode) {
	if len(previousNodes) == 0 {
		n.depth = 0
	}
	n.LR = previousNodes // 这里先不排序
	maxLevel := -1
	for _, node := range n.LR {
		if node.Depth() > maxLevel {
			maxLevel = node.Depth()
		}
	}
	maxLevel++
	n.depth = maxLevel
}
func (n *InstructionNode) SignToMiddle() {
	if !n.IsVisited() {
		panic("Node should be Visited!!!")
	}
	n.visited = Middle
}
func (n *InstructionNode) IsMiddle() bool {
	return n.visited == Middle
}
func (n *InstructionNode) Visit() {
	if n.IsVisited() {
		panic("This Node has been Visited!!!")
	}
	n.visited = Normal
}
func (n *InstructionNode) TryVisit() bool {
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
func (n *InstructionNode) IsVisited() bool {
	return n.visited != NotVisited
}
func (n *InstructionNode) AddDegree() {
	n.degree++
}
func (n *InstructionNode) SetSplit(s int) bool {
	if n.split != SPLIT_UNSET {
		panic("This Node's Split has been Set!!!")
	}
	n.split = s
	return true
}

// Ergodic 后续遍历
// todo 节点遍历逻辑，后序遍历
func (n *InstructionNode) Ergodic(b *Bucket) {
	// 1. 先遍历depth小的父节点,depth一样先遍历出度小的节点
	// 2. 每次在添加O后判断是否弹出bucket:
	//    (1) 如果还没满，添加O
	//    (2) 如果满了，留一个（这个待定）位置，判断下一个O是否也已经遍历完LR，如果不是则添加O并弹出Bucket，反之直接弹出Bucket

	// LR里的所有节点按照depth和出度排序
	sort := func(a LroNode, b LroNode) bool {
		if a.Depth() < b.Depth() {
			return true
		} else if a.Depth() == b.Depth() {
			return a.Degree() <= b.Degree()
		}
		return false
	}
	se := NewSortEngine(sort)
	n.LR = se.Sort(n.LR).([]LroNode)
	// 冒泡排序
	//for i := 0; i < len(n.LR); i++ {
	//	for j := 0; j < len(n.LR)-i-1; j++ {
	//		if !sort(n.LR[j], n.LR[j+1]) {
	//			n.LR[j], n.LR[j+1] = n.LR[j+1], n.LR[j]
	//		}
	//	}
	//}
	// 后续遍历
	for _, node := range n.LR {
		if !node.TryVisit() {
			//n.CheckMiddle()
			continue
		}
		node.Ergodic(b)
	}
	if n.O == FAKE_ROOT_ID {
		return
	}
	n.Visit()
	b.Add(n)
	for _, node := range n.LR {
		node.CheckMiddle(n.split, b)
	}
}
func (n *InstructionNode) CheckMiddle(split int, b *Bucket) {
	if n.split != split {
		n.visited = Middle
		b.UpdateSplitLevel(n.split, split)
	}
}

//func (n *LroNode) CheckFinish() bool {
//	for _, node := range n.LR {
//		if !node.Visit() {
//			return false
//		}
//	}
//}
