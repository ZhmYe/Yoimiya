package Streaming

type ConstraintGraph struct {
	roots []*ConstraintNode // 所有inputNode
	nodes []*ConstraintNode // 所有Instruction
}

func (g *ConstraintGraph) GetNode(id int) *ConstraintNode {
	if id < 0 {
		wireID := -id
		for wireID > len(g.roots) {
			g.roots = append(g.roots, NewConstraintNode(len(g.roots)+1, make([]int, 0)))
		}
		return g.roots[-id-1]
	}
	return g.nodes[id]
}
func (g *ConstraintGraph) Insert(iID int, previousIDs []int) {
	instruction := NewConstraintNode(iID, previousIDs)
	// 这里当检测到previousID < 0时，添加root
	for _, id := range previousIDs {
		previousNode := g.GetNode(id)
		//previousNode.NotRoot()
		previousNode.AddDegree() // 增加出度
		previousNode.InsertChild(instruction)
		instruction.UpdateDepth(previousNode.Depth())
		//previousNodes = append(previousNodes, previousNode)
	}
	g.nodes = append(g.nodes, instruction)
}
