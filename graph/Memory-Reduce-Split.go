package graph

import (
	"sort"
	"sync"
)

// 这里写切割电路的测试版本
// 电路是一个有向无环图，由一定数量的变量(variables)组成，这些variable的值需要被计算并保存
// 所有的variable的值最终会作为proof的一部分(*)
// 电路的有向性体现在，一些variable需要接收前置的variables的值作为输入，然后根据r1cs的a · b = c计算得到其自己的值
// 因此，前面的variable算完才能算后面的variable(**)
// 再来看电路的本质，是要得到一个A(1, w) · B(1, w) = C(1, w), 在代码里A，B, C是三个矩阵，而·是hadamard product
// 将hadamard product的每一维单独拎出来，就可以转化为计算a · b = c，这样的每一维的计算在代码中作为Instruction
// 每个Instruction中只有一个需要计算的值(output)，每个Instruction中包含了一定数量的Variable，称为Wire (***)
// 因此事实上，电路可以被认为是Instruction组成的有向无环图
// (在之前我写的代码中，我将这个Instruction组成的DAG进行了划分，组成了一个Stage-Instruction-Tree,在sit.go， 后续再考虑是切sit还是切原始的instruction-tree，无所谓)
// 这里最麻烦的地方在于，前面(*)提到过，所有的variable会作为proof的一部分，我们知道树形结构，到一定深度后，有些变量可能就没有后置节点了，因此理论上老说它的内存可以回收
// 但现在因为后续proof需要，因此它的内存不能丢，导致整个内存使用会随着wire的数量增加而增加
// 因此，目前山东大学一篇论文提出对电路进行切割，但其解法是代码层面切割（一个for 1-100的循环，改成两次，for 1-50, 50-100），不是通用解法
// 我们就是要做一个通用解法
// 本质上就是， 我们将原来的电路dag切成两部分跑
// 记原本的电路是y = F(x), 那么划分成两个电路y_1 = F_1(x_1), y = F_2(x_2, y_1), 其中x_1 ∩ x_2 = x，大概这样
// 但这里要考虑一个问题，就是我们生成多少个proof
// 如果还是生成一个Proof，前面说过所有variable还是要保存，因此切了和没切没啥区别，所以一定是生成多个proof
// 换句话说，切完以后不是跑一次，而是跑两次（n次）
// 也就是说比如我们得到两个电路c_1, c_2，是要将他按照原来的代码形式生成两次proof，这里就会出现一个问题
// 我们需要精准修改c_1,c_2的input和output，包括数量和连接，这部分后续我会看
// 你们现在先在这里，这个文件里，实现上述切割电路，本质上就是
// 如何在一个图G=(E,V)中选择一组顶点V‘(记与这些顶点相关的边为E')
// 使得G’=(E\E',V\V')可以被划分为两个互不相关的图G1,G2
// 其中V‘中的顶点两两互不相连
// 尽可能写一个类的形式，比如SplitEngine，接受一定形式的输入返回输出
// 这边不要用邻接矩阵表示图，我这里会有176w个wire。如果用邻接矩阵176w * 176w直接爆炸
// 尽量就按照这样的形式：我这个图，有多个根节点*root, 然后每个节点（Node）,是一个结构体， 有一个成员变量child（不止一个），就是类似链表的结构
// 可以参考stage.go的stage定义

func generateNewSIT(t *SITree, fatherMap map[int]map[int]bool, latterRoot []*Stage, childList map[int]bool) (*SITree, *SITree) {
	sortList := make([]int, 0)
	dummyStage := NewStage(-1)
	for _, stage := range t.root {
		dummyStage.AddChild(stage)
	}
	dummyStage.AddInstruction(-1, true)
	generateNewSITFoo(&sortList, make(map[int]bool), dummyStage)
	sortList = sortList[1:]
	idxMap := make(map[int]int)
	for i, idx := range sortList {
		idxMap[idx] = i
	}
	formerStages := make([]*Stage, 0)
	formerRoot := t.GetRootStages()
	latterStages := make([]*Stage, 0)
	latterRootMap := make(map[int]bool)
	for _, stage := range latterRoot {
		idx := stage.GetLastInstruction()
		stage.parent = 0
		latterRootMap[idx] = true
		for key := range fatherMap[idx] {
			t.GetStageByInstruction(key).DelChild(stage)
		}
	}
	for _, stage := range formerRoot {
		if latterRootMap[stage.GetLastInstruction()] {
			return t, nil
		}
	}
	allStages := t.GetStages()
	for _, stage := range allStages {
		if _, ok := childList[stage.GetLastInstruction()]; !ok {
			formerStages = append(formerStages, stage)
		} else {
			latterStages = append(latterStages, stage)
		}
	}
	sort.Slice(formerStages, func(i, j int) bool {
		return idxMap[formerStages[i].GetLastInstruction()] < idxMap[formerStages[j].GetLastInstruction()]
	})
	sort.Slice(latterStages, func(i, j int) bool {
		return idxMap[latterStages[i].GetLastInstruction()] < idxMap[latterStages[j].GetLastInstruction()]
	})
	formerSIT := NewSITree()
	latterSIT := NewSITree()
	for _, stage := range formerStages {

		if stage.GetCount() == 0 {
			formerSIT.appendRoot(stage)
		}
		formerSIT.appendStage(stage)
		formerSIT.instructions = append(formerSIT.instructions, stage.GetInstructions()...)
	}
	for _, stage := range latterStages {
		if stage.GetCount() == 0 {
			latterSIT.appendRoot(stage)
		}
		latterSIT.appendStage(stage)
		latterSIT.instructions = append(latterSIT.instructions, stage.GetInstructions()...)
	}
	return formerSIT, latterSIT
}

func generateNewSITFoo(sortList *[]int, visited map[int]bool, curStage *Stage) {
	if visited[curStage.GetLastInstruction()] {
		return
	}
	visited[curStage.GetLastInstruction()] = true
	for _, stage := range curStage.GetSubStages() {
		if _, ok := visited[stage.GetLastInstruction()]; ok {
			continue
		}
		generateNewSITFoo(sortList, visited, stage)
	}
	*sortList = append([]int{curStage.GetLastInstruction()}, *sortList...)
}

func computeSITStagesWeight(t *SITree) (weightMap map[int]float64, fatherMap map[int]map[int]bool) {
	weightMap = make(map[int]float64)
	fatherMap = make(map[int]map[int]bool)
	leafMap := make(map[int]bool)
	weightLatch := sync.Mutex{}
	wG := sync.WaitGroup{}
	for _, stage := range t.root {
		dfsSITStages1(stage, weightMap, fatherMap, leafMap, &weightLatch, 0)
	}

	for key := range leafMap {
		wG.Add(1)
		go func(key int) {
			dfsSITStages2(key, weightMap, fatherMap, &weightLatch, 0)
			wG.Done()
		}(key)
	}
	wG.Wait()
	return
}
func weightMapFixChild(t *SITree, weightMap map[int]float64, fatherMap map[int]map[int]bool, childList map[int]bool, doneMap map[int]bool, idx int) {
	if doneMap[idx] {
		return
	}
	curStage := t.GetStageByInstruction(idx)
	childrenStage := curStage.GetSubStages()
	if len(childrenStage) == 0 {
		childList[idx] = true
		delete(weightMap, idx)
		doneMap[idx] = true
		//delete(fatherMap, idx)
		return
	}
	wG := sync.WaitGroup{}
	for _, stage := range childrenStage {
		childIdx := stage.GetLastInstruction()
		if doneMap[childIdx] {
			continue
		}
		childList[childIdx] = true
		wG.Add(1)
		go func(childIdx int) {
			weightMapFixChild(t, weightMap, fatherMap, childList, doneMap, childIdx)
			wG.Done()
		}(childIdx)
	}
	wG.Wait()
	delete(weightMap, idx)
	doneMap[idx] = true
	//delete(fatherMap, idx)
}

func weightMapFixFather(weightMap map[int]float64, fatherMap map[int]map[int]bool, doneMap map[int]bool, idx int) {
	if doneMap[idx] {
		return
	}
	wG := sync.WaitGroup{}
	if _, ok := fatherMap[idx]; !ok {
		delete(weightMap, idx)
		doneMap[idx] = true
		return
	}
	fm := fatherMap[idx]
	for key := range fm {
		if fm[key] {
			if doneMap[key] {
				continue
			}
			wG.Add(1)
			go func(key int) {
				weightMapFixFather(weightMap, fatherMap, doneMap, key)
				wG.Done()
			}(key)
		}
	}
	wG.Wait()
	delete(weightMap, idx)
	doneMap[idx] = true
	//delete(fatherMap, idx)
}

func dfsSITStages1(stage *Stage, weightMap map[int]float64, fatherMap map[int]map[int]bool, leafMap map[int]bool, weightLatch *sync.Mutex, depth int) (score int) {
	if len(stage.GetSubStages()) == 0 {
		leafMap[stage.GetLastInstruction()] = true
		return depth
	}
	maxScore := -1
	fatherIdx := stage.GetLastInstruction()
	for _, child := range stage.GetSubStages() {
		score = dfsSITStages1(child, weightMap, fatherMap, leafMap, weightLatch, depth+1)
		if maxScore < score {
			maxScore = score
		}
		weightLatch.Lock()
		weightMap[fatherIdx] += float64(score) / float64(depth+1)
		weightLatch.Unlock()
		idx := child.GetLastInstruction()
		if _, ok := fatherMap[idx]; !ok {
			fatherMap[idx] = make(map[int]bool)
		}
		fatherMap[idx][fatherIdx] = true
	}
	if (depth >> 1) > maxScore {
		return 2 * (maxScore - depth)
	} else {
		return 2 * depth
	}
}

func dfsSITStages2(fatherIdx int, weightMap map[int]float64, childMap map[int]map[int]bool, weightLatch *sync.Mutex, depth int) (score int) {
	if _, ok := childMap[fatherIdx]; !ok {
		return depth
	}
	minScore := int(^uint(0) >> 1)
	for childIdx := range childMap[fatherIdx] {
		score = dfsSITStages2(childIdx, weightMap, childMap, weightLatch, depth+1)
		if minScore > score {
			minScore = score
		}
		weightLatch.Lock()
		weightMap[fatherIdx] += float64(score) / float64(depth+1)
		weightLatch.Unlock()
	}
	if (depth >> 1) > minScore {
		return minScore - depth
	} else {
		return depth
	}
}
