package graph

import "sync"

type Stage struct {
	id           int        // 用于标识stage, 存储父stage、子stage时需要
	Instructions []int      // 每个stage分为两阶段执行，首先执行一系列串行的Instruction，然后出现宽依赖后并行执行子stage
	child        []*Stage   // 子Stage,并行执行
	parent       []*Stage   // 父Stage，可能不止一个
	count        int        // 计数器，用于计算当前stage是否可以执行
	mutex        sync.Mutex // 锁count，可能会有并行的父stage修改当前stage
}

// WakeUp 父stage在并行执行各个子stage时尝试call这些子stage，调用这些stage的wakeup函数，使得其count++
func (s *Stage) WakeUp() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count++
	// 所有父stage的前置计算运行完成
	if s.count == len(s.parent) {
		s.Run()
	}
}

// Call 尝试执行子Stage
func (s *Stage) Call() {
	for _, subStage := range s.child {
		subStage.WakeUp()
	}
}
func (s *Stage) Run() {
	return
}

// AddChild 添加子Stage
func (s *Stage) AddChild(child *Stage) {
	s.child = append(s.child, child)
}

// AddParent 添加父Stage
func (s *Stage) AddParent(parent *Stage) {
	s.parent = append(s.parent, parent)
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
func NewStage(id int, instructions []int) *Stage {
	stage := new(Stage)
	stage.id = id
	stage.count = 0
	stage.Instructions = instructions
	stage.child = make([]*Stage, 0)
	stage.parent = make([]*Stage, 0)
	return stage
}
