package Pipeline

import "Yoimiya/frontend"

// Coordinator
// 在原本的Parallel中，一个Slot会被一个完整的电路填充，一共有maxParallel个Slot同时运行
// Pipeline中，将一个Slot中的任务，通过split的方式达到Misaligned Parallel的效果
type Coordinator struct {
	tasks       []frontend.PackedLeafInfo // 每个任务用它assignment所对应的packedLeafInfo表示
	nbSplit     int                       // split数量
	maxParallel int                       // slot数量
	pcs         []PackedConstraintSystem  // split电路cs与其对应的pk,vk
}
