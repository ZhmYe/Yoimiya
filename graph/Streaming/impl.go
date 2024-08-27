package Streaming

type YoimiyaSplitEngine struct {
	orderer     Orderer         // 用于将输入的图进行排序然后传输到Partitioner里
	partitioner Partitioner     // 用于对Streaming进行划分
	graph       ConstraintGraph // ccs对应的DAG
}

func (se *YoimiyaSplitEngine) GetLayersInfo() [4]int {
	return [4]int{0, 0, 0, 0}
}

//CheckAndGetSubCircuitStageIDs() ([]int, []int)

func (se *YoimiyaSplitEngine) Insert(iID int, previousIDs []int) {
	se.graph.Insert(iID, previousIDs)
}
func (se *YoimiyaSplitEngine) AssignLayer(cut int) {

}

//GetInstructionIdsFromStageIDs(ids []int) []int

func (se *YoimiyaSplitEngine) GetSubCircuitInstructionIDs() [][]int {

}
func (se *YoimiyaSplitEngine) GetStageNumber() int {
	return se.partitioner.GetPartitionNumber()
}
func (se *YoimiyaSplitEngine) GetMiddleOutputs() map[int]bool {

}
func (se *YoimiyaSplitEngine) GetAllInstructions() []int {

}
func (se *YoimiyaSplitEngine) IsMiddle(iID int) bool {

}
func (se *YoimiyaSplitEngine) GenerateSplitWitness() [][]int {

}
