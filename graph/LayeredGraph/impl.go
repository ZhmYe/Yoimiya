package LayeredGraph

// LayeredGraph PackedLevel用于记录某一instruction自身所在的LEVEL，以及它的子节点最大深度
// 在PackedLevel的基础上加上逆向连接
// 也就是分层图，只不过这个分层图是逆向的
// 逆向的layer是原先level的等价形式，拓扑排序可逆
// 逆向的原因是，倒着遍历可以得到某个instruction的所有父节点，反之如果顺序层次遍历，加入某个节点的时候没法把它所有父节点加入到备选
type LayeredGraph struct {
	//root []*Node // 没有父节点的节点，也就是原本的DAG中的叶子节点
	Levels [][]int // 同core.go中的LEVELs
	nodes  map[int]*Node
}

func NewLayeredGraph() *LayeredGraph {
	return &LayeredGraph{
		Levels: make([][]int, 0),
		nodes:  make(map[int]*Node),
	}
}
func (g *LayeredGraph) GetNodeByiID(iID int) *Node {
	node, exist := g.nodes[iID]
	if !exist {
		panic("No such Node!!!")
	} else {
		return node
	}
}
func (g *LayeredGraph) GetRootNodes() []*Node {
	// 没有父节点的节点，不一定就是最后一层
	//root := g.Levels[len(g.Levels)-1]
	result := make([]*Node, 0)
	for _, level := range g.Levels {
		for _, id := range level {
			node := g.nodes[id]
			if node.IsRoot() {
				result = append(result, node)
			}
		}
	}
	return result
}

// Insert 插入一个新的节点，更新连接关系，并维护level
// 这里previousID是原始图中的前置instruction，在当前的layerGraph中应该是Node的子节点，但是他们先被插入
func (g *LayeredGraph) Insert(iID int, previousIDs []int) {
	node := NewNode(iID)
	g.nodes[iID] = &node
	maxLevel := -1
	for _, id := range previousIDs {
		child := g.GetNodeByiID(id)
		child.NotRoot()
		node.AddChild(child)
		if child.level > maxLevel {
			maxLevel = child.level
		}
	}
	maxLevel++
	// 得到Node所在level
	if maxLevel >= len(g.Levels) {
		g.Levels = append(g.Levels, []int{iID})
	} else {
		// 将iID添加到level
		g.Levels[maxLevel] = append(g.Levels[maxLevel], iID)
	}
	node.AssignLevel(maxLevel)
}
func (g *LayeredGraph) GetLayersInfo() [4]int {
	return [4]int{0, 0, 0, 0}
}

// AssignLayer 划分split
// todo
// （倒着）遍历整个分层图，维护一个队列，每次从队列中取出出度最小的（子节点最少的），然后将其所有子节点放入队列
// 出度最小，说明取出该节点后，上面所需的middle越少
// 每次取出n/cut个节点
// 这里队列就先简单的维护一个数组，然后遍历一遍，不排序
func (g *LayeredGraph) AssignLayer(cut int) {
	queue := make([]*Node, 0)
	subGraphIndex := 0
	nbInstruction := g.NbInstruction()
	threshold := nbInstruction / cut
	selected := make([]*Node, 0) // 某个子图

	pop := func(queue []*Node) (*Node, []*Node) {
		// 取出队列中出度最小的节点，然后返回新的队列
		minDegree := -1
		split := -1
		for index, node := range queue {
			if minDegree == -1 {
				split = index
				minDegree = node.GetDegree()
			} else {
				if node.GetDegree() < minDegree {
					minDegree = node.GetDegree()
					split = index
				}
			}
		}
		selectNode := queue[split]
		queue = append(queue[:split], queue[split+1:]...)
		return selectNode, queue
	}
	// 这里写成批量的形式，因为层次遍历不止一个子节点
	push := func(queue []*Node, predict int, nodes ...*Node) []*Node {
		for _, node := range nodes {
			// 如果节点已经被访问过了，那么有两种可能
			if node.IsVisited() {
				// case1: 节点还没有分配split，说明有原图至少两个子节点
				if !node.HasBeenSplit() {
					// case1.1
					// 如果node原本的predict值和现在的predict不一样，说明节点的子节点在不同的子图中
					// 并且前一个子图已经满了，后面的节点还可能加入到后一个子图中
					// 那么将节点直接select到G1中
					if node.predict != predict && node.predict != SignMiddle {
						//selected = append(selected, node)
						node.AssignSplit(node.predict)
						// 此时前面的node.HashBeenSplit不可能再次判断成功
						// 因此不会有后续其他的处理
						// todo 这样threshold的逻辑要改
					}
					// case1.2 如果node.predict == node.predict',那么就是放在同一个子图里
					// 换句话说，两个子节点在同一个子图里，不做额外的处理

					// 这里节点每次被尝试push都会出现这段逻辑
					// 因此这里事实上最多触发一次逻辑
					// 也就是说，在某一个图Gk中的子节点a,认为自己的父节点b应该被放到Gk中，但因为贪心选择问题
					// 导致b没有被放到Gk里，我们先考虑b的祖先节点c是否可能被放到Gk甚至之前的图Gm,m<=k
					// 这是可能的
				} else {
					// case2: 节点已经被分配了split
					// 也就是说节点已经被分配到了前面的子图G1里，显然predict != split
					// 那么形成了一个顺序G1->G2
					// 节点被split之前出现的逻辑应该在case1中被考虑完整
					// todo
					// case2.1 如果已经有顺序G2->G1
					// 这里将该节点单独取出作为新的子图G3，这样出现顺序G3->G2
					// 首先该节点不可能是根节点，不然不可能会出现有节点会尝试将它push的情况
					// 那么既然该节点被加入到了G1中，说明存在原图中的子节点也在G1，即有顺序G3->G1

					// 遍历该节点的所有子节点（原图的父节点），判断是否有子节点已经被划分为G1
					// 如果有，那么出现顺序G1->G3，那么将这些子节点划分为G3,同时继续递归判断这些子节点的子节点
					// 并将那些没有划分的子节点的predict修改为G3

				}
			} else {
				// 如果还没访问过，那就正常将节点加入到队列里，并做predict
				queue = append(queue, node)
				node.Visit()
				node.Predict(predict)
			}
		}
		return queue
	}
	// 这里的预测值为-1
	queue = push(queue, -1, g.GetRootNodes()...) // 先将所有的root Node加入到队列
	assign := func(nodes []*Node) {
		for _, node := range nodes {
			node.AssignSplit(cut - subGraphIndex - 1) // 这里我们是倒着遍历的，因此第i个子图应该是第cut-i个split
			node.CheckMiddle()
		}
	}
	for {
		var node *Node
		// 取出出度最小的节点
		node, queue = pop(queue)
		// 将所有“父”节点加入到队列中
		queue = push(queue, cut-subGraphIndex-1, node.GetChildren()...)

		// 将节点加入到子图中
		selected = append(selected, node)
		// todo 如何划分middle
		// 每个节点加入的时候，把它的子节点未来可能属于的Split做一个预测
		// case 1: 如果该子节点有多个父节点，（原图中有多个子节点），那么会被预测多次，则后续几次预测时如果结果和前面不一样，直接置为Middle
		// case 2: 如果该子节点只有一个父节点，那么只会被预测一次，然后在真实加入split的时候，将实际的split和预测值比对，如果不一样置为Middle
		// todo 这里暂时没有处理cut=1，但可以放到外面去用noSplit处理
		if len(selected) == threshold {
			assign(selected)
			selected = make([]*Node, 0) // 清空selected
			subGraphIndex++
			//fmt.Println(subGraphIndex, len(queue))
			if subGraphIndex+1 == cut {
				// 最后一个图，阈值需要改变
				threshold = nbInstruction - subGraphIndex*threshold
				//selected = append(selected, queue...)
				//assign(selected)
				//break
			}
			if subGraphIndex == cut {
				if len(queue) != 0 {
					panic("not all elements are selected!!!")
				}
				break
			}
		}
	}
}
func (g *LayeredGraph) GetSubCircuitInstructionIDs() [][]int {
	result := make([][]int, 0)
	// 这里的LEVEL还是PackedLevel一样，是有原图的拓扑排序的
	for _, level := range g.Levels {
		for _, id := range level {
			node := g.nodes[id]
			split := node.split
			if split >= len(result) {
				for i := len(result); i <= split; i++ {
					result = append(result, make([]int, 0))
				}
			}
			result[split] = append(result[split], id)
		}
	}
	return result
}
func (g *LayeredGraph) GetStageNumber() int {
	return len(g.Levels)
}
func (g *LayeredGraph) GetMiddleOutputs() map[int]bool {
	result := make(map[int]bool)
	for _, level := range g.Levels {
		for _, id := range level {
			node := g.nodes[id]
			if node.IsMiddle() {
				result[id] = true
			}
		}
	}
	return result
}
func (g *LayeredGraph) GetAllInstructions() []int {
	iIDs := make([]int, 0)
	for _, level := range g.Levels {
		for _, id := range level {
			iIDs = append(iIDs, id)
		}
	}
	return iIDs
}
func (g *LayeredGraph) NbInstruction() int {
	return len(g.nodes)
}
func (g *LayeredGraph) IsMiddle(iID int) bool {
	node := g.nodes[iID]
	return node.IsMiddle()
}
