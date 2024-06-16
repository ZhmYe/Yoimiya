package LRO_Tree

import "fmt"

type LroTree struct {
	// 这里的iID是依次加一的，可以对应下标
	nodes  []*LroNode // todo 这里是否需要存这么多node，然后找到root?
	bucket *Bucket
}

func NewLroTree() *LroTree {
	return &LroTree{nodes: make([]*LroNode, 0), bucket: NewBucket()}
}
func (t *LroTree) GetNode(iID int) *LroNode {
	return t.nodes[iID]
}

// GetLeafNode 限制只有一个叶子节点
// 原本可能有若干个没有后续节点的，将他们统一指向一个伪造的叶子节点
func (t *LroTree) GetLeafNode() *LroNode {
	Leafs := make([]*LroNode, 0)
	for _, node := range t.nodes {
		if node.IsRoot() {
			Leafs = append(Leafs, node)
		}
	}
	// 不止一个叶子节点
	if len(Leafs) != 1 {
		fakeLeaf := NewLroNode(FAKE_ROOT_ID)
		fakeLeaf.SetLR(Leafs)
		return fakeLeaf
	} else {
		return Leafs[0]
	}
}
func (t *LroTree) Insert(iID int, previousIDs []int) {
	node := NewLroNode(iID)
	previousNodes := make([]*LroNode, 0)
	for _, id := range previousIDs {
		previousNode := t.GetNode(id)
		previousNode.NotRoot()
		previousNode.AddDegree() // 增加出度
		previousNodes = append(previousNodes, previousNode)
	}
	node.SetLR(previousNodes)
	t.nodes = append(t.nodes, node)
}
func (t *LroTree) AssignLayer(cut int) {
	threshold := RoundUpSplit(t.NbInstruction(), cut) // 每一份的阈值
	t.bucket.SetThreshold(threshold)
	// todo 先写后续遍历的逻辑
	Leaf := t.GetLeafNode()
	Leaf.Ergodic(t.bucket)
	fmt.Println(t.bucket.split)
}
func (t *LroTree) IsMiddle(iID int) bool {
	// todo 怎么判断节点是否是middle
	node := t.GetNode(iID) // 获取节点
	return node.IsMiddle()
}
func (t *LroTree) GetAllInstructions() []int {
	iIDs := make([]int, 0)
	for i, node := range t.nodes {
		iIDs[i] = node.O
	}
	return iIDs
}
func (t *LroTree) GetMiddleOutputs() map[int]bool {
	return make(map[int]bool)
}
func (t *LroTree) GetStageNumber() int {
	return t.bucket.split
}

// GetSubCircuitInstructionIDs todo
func (t *LroTree) GetSubCircuitInstructionIDs() [][]int {
	result := make([][]int, 0)
	nbSplit := t.bucket.split + 1
	for i := 0; i < nbSplit; i++ {
		result = append(result, make([]int, 0))
	}
	// 这里的iID是依次加一的，可以对应下标，拓扑有序
	for _, node := range t.nodes {
		split := node.split
		result[split] = append(result[split], node.O)
	}
	return result
}
func (t *LroTree) GetLayersInfo() [4]int {
	return [4]int{0, 0, 0, 0}
}
func (t *LroTree) NbInstruction() int {
	return len(t.nodes)
}
