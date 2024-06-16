package LayeredGraph

const SignMiddle int = -2

// Node 图上的节点，表示一个instruction，记录其所有“父节点”
// 图中每个level存放这些Node
type Node struct {
	iID int // instruction id
	//piIDs []int // 所有父节点的id，在图上是子节点
	LRNodes []*Node // 所有前置instruction，在图上是子节点
	level   int     // level记录在这里，避免外面需要使用Map
	split   int     // 记录Node在哪个split里
	isRoot  bool    // 记录是否为root
	//middle  bool    // 记录Node是否为middle
	predict int  // 对于split的预测值，用-1表示未设置，用SignMiddle表示middle
	visited bool // 表示节点是否被选过了
	depth   int  // 记录节点的最大深度
	// todo 如果判断是否是middle
}

func NewNode(iID int) Node {
	return Node{
		iID:     iID,
		LRNodes: make([]*Node, 0),
		level:   -1,
		split:   -1,
		predict: -1,
		visited: false,
		isRoot:  true,
		depth:   -1,
		//deepest: -1,
	}
}
func (n *Node) IsRoot() bool {
	return n.isRoot
}
func (n *Node) NotRoot() {
	n.isRoot = false
}
func (n *Node) HasBeenSplit() bool {
	return n.split != -1
}

//	func (n *Node) SignMiddle() {
//		n.middle = true
//	}
func (n *Node) CheckMiddle() {
	if n.predict == SignMiddle {
		return
	}
	// 每个节点都会有个predict，除非它是root,root不可能作为middle
	if n.predict == -1 {
		//panic("Predict Not Set!!!")
		return

	}
	if n.predict != n.split {
		n.predict = SignMiddle
	}
}
func (n *Node) IsVisited() bool {
	return n.visited
}
func (n *Node) Visit() {
	n.visited = true
}
func (n *Node) Predict(s int) {
	if n.predict == -1 {
		// 还没有设置预测值，那么直接赋值即可
		n.predict = s
	} else {
		// 如果当前节点已经被判断为一定是Middle，那么不用做任何判断了
		if n.predict == SignMiddle {
			return
		}
		// 判断现在的predict和s是否一致，如果不一样，则将n是middle
		if n.predict != s {
			n.predict = SignMiddle
		}
	}
}
func (n *Node) IsMiddle() bool {
	return n.predict == SignMiddle
}
func (n *Node) AddChild(c *Node) {
	n.LRNodes = append(n.LRNodes, c)

	// c多了一个父节点，更新deepest
	c.UpdateDeepest(c.GetDepth())
}
func (n *Node) AssignSplit(s int) {
	if n.split != -1 {
		panic("Node Split has been set!!!")
	}
	n.split = s
}
func (n *Node) GetChildren() []*Node {
	return n.LRNodes
}
func (n *Node) AssignLevel(level int) {
	if n.level != -1 {
		panic("level has been assigned!!!")
	}
	n.level = level
}

// GetDegree 出度，也就是子节点的数量
func (n *Node) GetDegree() int {
	return len(n.LRNodes)
}

// UpdateDeepest 更新节点深度
func (n *Node) UpdateDeepest(depth int) {
	if depth == -1 {
		panic("Level Unset!!!")
	}
	if n.depth == -1 {
		// 如果还没有设置过deepest
		n.depth = depth + 1
	} else {
		if depth+1 >= n.depth {
			n.depth = depth + 1
		}
	}
}
func (n *Node) GetDepth() int {
	return n.depth
}