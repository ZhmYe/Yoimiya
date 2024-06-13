package PartitionGraph

type PartitionGraph struct {
	// todo 这里的iID应该是顺序的
	nodes      map[int]*Node // iID -> Node，Node中记录了其所在的partition和level
	partitions []*Partition  // 所有分区,PID对应数组下标
	root       []int         // 所有根分区
}

func NewPartitionGraph() *PartitionGraph {
	//rootPartition := NewPartition()
	//rootPartition.SetPID(0) // 根分区
	//rootPartition.AssignRoot()
	g := PartitionGraph{
		nodes:      make(map[int]*Node),
		partitions: make([]*Partition, 0),
	}
	return &g
}
func (g *PartitionGraph) AddNode(iID int) *Node {
	node := NewNode(iID)
	g.nodes[iID] = &node
	return g.nodes[iID]
}
func (g *PartitionGraph) getNodePID(iID int) int {
	node := g.nodes[iID]
	if node.partition == PID_UNSET {
		panic("Node Partition Not Set!!!")
	}
	return node.partition
}
func (g *PartitionGraph) AddRootPartition(iID int) {
	partition := NewPartition()
	partition.AssignRoot()
	partition.SetPID(len(g.partitions))
	g.partitions = append(g.partitions, partition)
	partition.AddLevel([]int{iID})
	g.nodes[iID].AssignPartitionAndLevel(partition.pID, 0)
}

// Insert 插入一个Instruction
func (g *PartitionGraph) Insert(iID int, previousIDs []int) {
	// 首先创建一个Node
	node := g.AddNode(iID)
	if len(previousIDs) == 0 {
		// 没有父节点，也就是根节点，根节点单独维护一个分区，并将分区标记为root
		g.AddRootPartition(iID)
		return
	}
	// 父节点所在分区
	previousPID := make(map[int]bool)
	tmp := -1
	levels := make([]int, 0)
	for _, id := range previousIDs {
		//tmp = g.getNodePID(id)
		pNode := g.nodes[id]
		tmp = pNode.partition
		previousPID[pNode.partition] = true
		levels = append(levels, pNode.level)
	}
	// 父节点在一个分区里
	if len(previousPID) == 1 {
		// 将节点添加到父节点所在分区
		// 处理分区可能分裂的逻辑
		partition := g.partitions[tmp]
		isSplit, level, newPartitions := partition.Insert(iID, levels)

		if !isSplit {
			// 如果不需要分裂，那么节点就被加入到了partition中，同时返回了其所在level
			node.AssignPartitionAndLevel(partition.pID, level)
		} else {
			for _, newPartition := range newPartitions {
				newPartition.SetPID(len(g.partitions))
				g.partitions = append(g.partitions, newPartition) // 添加新的分区
				assign := func(partition *Partition) {
					for lID, l := range partition.levels {
						for _, id := range l {
							assignNode := g.nodes[id]
							assignNode.AssignPartitionAndLevel(partition.pID, lID)
						}
					}
				}
				// 将新分区中的每个节点的信息重新赋予
				// todo 这里其实可以放到Partition内部的split逻辑，前提是分裂出的Partition PID可知
				assign(newPartition)
			}
		}
	} else {
		// 如果父节点不止一个分区
		// 那么需要新起一个分区
		partition := NewPartition()
		partition.SetPID(len(g.partitions))
		g.partitions = append(g.partitions, partition)
		partition.AddLevel([]int{iID})
		g.nodes[iID].AssignPartitionAndLevel(partition.pID, 0)
	}
}
func (g *PartitionGraph) AssignLayer(cut int) {

}
func (g *PartitionGraph) IsMiddle(iID int) bool {
	return false
}
func (g *PartitionGraph) GetAllInstructions() []int {
	return make([]int, 0)
}
func (g *PartitionGraph) GetMiddleOutputs() map[int]bool {
	return make(map[int]bool)
}
func (g *PartitionGraph) GetStageNumber() int {
	return 0
}
func (g *PartitionGraph) GetSubCircuitInstructionIDs() [][]int {
	return make([][]int, 0)
}
func (g *PartitionGraph) GetLayersInfo() [4]int {
	return [4]int{0, 0, 0, 0}
}
