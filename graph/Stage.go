package graph

import (
	"sync"
)

type Stage struct {
	id           int      // 用于标识stage, 存储父stage、子stage时需要
	Instructions []int    // 每个stage分为两阶段执行，首先执行一系列串行的Instruction，然后出现宽依赖后并行执行子stage
	child        []*Stage // 子Stage,并行执行
	parent       int      // 用于代替count
	//count        int        // 计数器，用于计算当前stage是否可以执行
	mutex sync.Mutex // 锁count，可能会有并行的父stage修改当前stage
}

func (s *Stage) GetID() int {
	return s.id
}
func (s *Stage) SetID(id int) {
	s.id = id
}

// WakeUp 父stage在并行执行各个子stage时尝试call这些子stage，调用这些stage的wakeup函数，使得其count++
func (s *Stage) WakeUp() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//fmt.Println(s.id)
	// RootStage
	if s.parent == 0 {
		return true
	} else {
		s.parent--
		if s.parent == 0 {
			return true
		}
	}
	return false
}

// AddChild 添加子Stage
func (s *Stage) AddChild(child *Stage) {
	s.child = append(s.child, child)
}

// AddParent 添加父Stage
func (s *Stage) AddParent(parent *Stage) {
	s.parent++
}

// AddInstruction 添加Instruction
func (s *Stage) AddInstruction(id int, reverse bool) {
	// reverse表示从头插入，真实使用时确实需要从头插入，因为在划分Stage的时候是从后往前遍历的
	if reverse {
		s.Instructions = append([]int{id}, s.Instructions...)
	} else {
		s.Instructions = append(s.Instructions, id)
	}
}
func (s *Stage) BatchAddInstruction(ids []int, reverse bool) {
	if reverse {
		s.Instructions = append(ids, s.Instructions...)
	} else {
		s.Instructions = append(s.Instructions, ids...)
	}
}
func (s *Stage) CutInstruction(index int) {
	s.Instructions = s.Instructions[:index]
}

//// GetParentIDs 返回所有父Stage的ID
//func (s *Stage) GetParentIDs() (ids []int) {
//	for _, stage := range s.parent {
//		ids = append(ids, stage.id)
//	}
//	return ids
//}

// GetChildIDs 返回所有子Stage的ID
func (s *Stage) GetChildIDs() (ids []int) {
	for _, stage := range s.child {
		ids = append(ids, stage.id)
	}
	return ids
}

// GetInstructions 返回所有Instruction
func (s *Stage) GetInstructions() []int {
	return s.Instructions
}
func (s *Stage) GetCount() int {
	return s.parent
}

func (s *Stage) GetSubStages() []*Stage {
	return s.child
}
func NewStage(id int, instructions ...int) *Stage {
	stage := new(Stage)

	//stage.count = 0
	stage.Instructions = make([]int, 0)
	for _, id := range instructions {
		stage.Instructions = append(stage.Instructions, id)
	}
	//stage.Instructions = instructions
	stage.child = make([]*Stage, 0)
	stage.parent = 0
	return stage
}
func (s *Stage) GetLastInstruction() int {
	return s.Instructions[len(s.Instructions)-1]
}
func (s *Stage) RemoveAllChild() {
	s.child = make([]*Stage, 0)
}
func (s *Stage) InheritChild(child []*Stage) {
	for _, child := range child {
		s.child = append(s.child, child)
	}
}
