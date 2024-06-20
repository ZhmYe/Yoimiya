package LRO_Tree

import "fmt"

type LroTree struct {
	// 这里的iID是依次加一的，可以对应下标
	instructions []*InstructionNode // instruction对应下标
	inputs       []*InputNode       // 这里用wireID来对应下标，在原电路生成的时候根据public, private的方式应该得到了前面的wireID
	bucket       *Bucket
}

func NewLroTree() *LroTree {
	return &LroTree{instructions: make([]*InstructionNode, 0), bucket: NewBucket()}
}
func (t *LroTree) GetInstruction(iID int) *InstructionNode {

	if iID >= len(t.instructions) {
		panic("instruction hasn't been insert!!!")
	}
	return t.instructions[iID]
}
func (t *LroTree) GetInput(wireID int) *InputNode {
	// todo 这里还不确定wireID是否是连续的
	for wireID > len(t.inputs) {
		t.inputs = append(t.inputs, NewInputNode(len(t.inputs)-1))
	}
	return t.inputs[wireID-1]
}

// GetLeafNode 限制只有一个叶子节点
// 原本可能有若干个没有后续节点的，将他们统一指向一个伪造的叶子节点
func (t *LroTree) GetLeafNode() *InstructionNode {
	Leafs := make([]*InstructionNode, 0)
	for _, node := range t.instructions {
		if node.IsRoot() {
			Leafs = append(Leafs, node)
		}
	}
	// 不止一个叶子节点
	if len(Leafs) != 1 {
		fakeLeaf := NewInstructionNode(FAKE_ROOT_ID)
		var tmp []LroNode
		for len(tmp) < len(Leafs) {
			tmp = append(tmp, Leafs[len(tmp)])
		}
		fakeLeaf.SetLR(tmp)
		return fakeLeaf
	} else {
		return Leafs[0]
	}
}
func (t *LroTree) GetNode(id int) LroNode {
	if id < 0 {
		return t.GetInput(-id)
	}
	return t.GetInstruction(id)
}
func (t *LroTree) InsertInput(wireID int) {
	t.inputs = append(t.inputs, NewInputNode(wireID))
}
func (t *LroTree) Insert(iID int, previousIDs []int) {
	instruction := NewInstructionNode(iID)
	previousNodes := make([]LroNode, 0)
	for _, id := range previousIDs {
		previousNode := t.GetNode(id)
		previousNode.NotRoot()
		previousNode.AddDegree() // 增加出度
		previousNodes = append(previousNodes, previousNode)
	}
	instruction.SetLR(previousNodes)
	t.instructions = append(t.instructions, instruction)
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
	for i, node := range t.instructions {
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

// GetSubCircuitInstructionIDs 得到子电路iIDs
func (t *LroTree) GetSubCircuitInstructionIDs() [][]int {
	result := make([][]int, 0)
	nbSplit := t.bucket.split + 1
	for i := 0; i < nbSplit; i++ {
		result = append(result, make([]int, 0))
	}
	// 这里的iID是依次加一的，可以对应下标，拓扑有序
	for _, node := range t.instructions {
		split := node.split
		result[split] = append(result[split], node.O)
	}
	return result
}
func (t *LroTree) GetLayersInfo() [4]int {
	return [4]int{0, 0, 0, 0}
}
func (t *LroTree) NbInstruction() int {
	return len(t.instructions)
}

// GenerateSplitWitness 这里通过InputNode的split得到每个split需要哪些Input，从而得到每个split的witness
// todo 后面还要根据这个[][]int得到具体的witness
func (t *LroTree) GenerateSplitWitness() [][]int {
	witnessSplits := make([][]int, 0)
	for index, inputNode := range t.inputs {
		for _, split := range inputNode.GetSplits() {
			for split > len(witnessSplits) {
				witnessSplits = append(witnessSplits, make([]int, 0))
			}
			witnessSplits[split] = append(witnessSplits[split], index+1)
		}
	}
	return witnessSplits
}
