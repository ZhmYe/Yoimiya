package graph

// 这里是对SIT.go的补充，为了避免Merge

// GetMiddleOutputs 返回所有Middle的Stage里最后一个Instruction
// todo 前面的Instruction里的Wire会不会也是Output?
func (t *SITree) GetMiddleOutputs() map[int]bool {
	result := make(map[int]bool)
	middleStage := t.GetMiddleStage()
	for _, id := range middleStage {
		stage := t.GetStageByIndex(id)
		result[stage.GetLastInstruction()] = true
	}
	return result
}
