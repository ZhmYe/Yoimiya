package graph

import (
	"fmt"
	"sync"
	"time"
)

const NoParent = -1

type SplitEngine struct {
	splitTime time.Duration        // 记录切分时间
	score     float64              // 暂定，记录切分好坏程度
	parentMap map[int]map[int]bool // 记录父节点
	sit       *SITree
	// todo 后续考虑之间放到stage内部？
	//depth     map[int]int // 记录深度
	//maxDepth  map[int]int // 记录最大深度
	//weightMap map[int]float64
}

func NewSplitEngine(sit *SITree) *SplitEngine {
	s := new(SplitEngine)
	s.splitTime = time.Duration(0)
	s.score = 0
	s.parentMap = make(map[int]map[int]bool)
	//s.depth = make(map[int]int)
	//s.maxDepth = make(map[int]int)
	//s.weightMap = make(map[int]float64)
	s.sit = sit
	return s
}

// Split 输入sit,将其划分为多个部分，每个部分用具有拓扑序的iID数组表示
// n表示划分为多少个部分，目前先不考虑 todo
func (s *SplitEngine) Split(n int) (*SITree, *SITree) {
	startTime := time.Now()
	ret := make([]*Stage, 0)
	//for _, rootStage := range s.sit.GetRootStages() {
	//	s.dfs(rootStage, 1, -1)
	//}
	//s.computeWeightMap()
	childList := make(map[int]bool)
	weightMap := s.GetWeightMap()
	fmt.Println("Score Compute Finished....")
	fmt.Println(weightMap)
	//fatherMap := s.GetParentMap()
	for len(weightMap) != 0 {
		Pos := -1
		Score := -1.0
		for k, v := range weightMap {
			if Score < v {
				Pos = k
				Score = v
			}
		}
		targetStage := s.sit.GetStageByInstruction(Pos)
		ret = append(ret, targetStage)
		childList[targetStage.GetLastInstruction()] = true
		for _, stage := range targetStage.GetSubStages() {
			childIdx := stage.GetLastInstruction()
			childList[childIdx] = true
			weightMapFixChild(s.sit, weightMap, childList, childIdx)
		}
		weightMapFixFather(s.sit, weightMap, &sync.Map{}, Pos)
	}
	fmt.Println("Try generate new SIT.")
	s.splitTime = time.Since(startTime)
	s.score = 0
	return generateNewSIT(s.sit, ret, childList)
}

//	GetWeightMap func (s *SplitEngine) computeWeightMap() map[int]float64 {
//		// 这里直接用depth
//		for id, depth := range s.depth {
//			s.weightMap[id] = float64((s.maxDepth[id] - depth) * (depth - 1))
//		}
//		return s.GetWeightMap()
//	}
func (s *SplitEngine) GetWeightMap() map[int]float64 {
	return s.sit.GetStageScore()
}

//	func (s *SplitEngine) GetParentMap() map[int]map[int]bool {
//		return s.parentMap
//	}
//func (s *SplitEngine) dfs(stage *Stage, depth int, parentID int) int {
//	_, exist := s.depth[stage.GetID()]
//	if !exist {
//		s.depth[stage.GetID()] = depth // 初始化depth
//		s.maxDepth[stage.GetID()] = depth
//		s.parentMap[stage.GetID()] = make(map[int]bool)
//	} else {
//		// 判断深度是否可能超过当前的深度，如果不是，那么无需更新，也无需继续递归，剪枝
//		if s.depth[stage.GetID()] >= depth {
//			return s.maxDepth[stage.GetID()]
//		}
//	}
//	if parentID != NoParent {
//		s.parentMap[stage.GetID()][parentID] = true
//	}
//	maxD := s.maxDepth[stage.GetID()]
//	flag := false // 用于表示是否需要更新depth
//	// 遍历所有子stage，获得最新的最长路径
//	for _, child := range stage.GetSubStages() {
//		if tmp := s.dfs(child, depth+1, stage.GetID()); tmp > maxD {
//			maxD = tmp
//			flag = true
//		}
//	}
//	if flag {
//		s.depth[stage.GetID()] = depth
//		s.maxDepth[stage.GetID()] = maxD
//	}
//	return s.maxDepth[stage.GetID()]
//}
