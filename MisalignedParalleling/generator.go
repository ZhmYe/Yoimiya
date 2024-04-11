package MisalignedParalleling

import (
	"Yoimiya/frontend"
)

// 这里我们生成n个一模一样的电路, 可以是输入不同的，但结构一样的
// 这里简单起见就直接n个一模一样的输入

// Generator 生成Task
type Generator struct {
	nbTask int // 产生的task数量
}

func (g *Generator) generate(assignmentInterface func() frontend.Circuit) []*Task {
	tasks := make([]*Task, 0)
	for i := 0; i < g.nbTask; i++ {
		assignment := assignmentInterface()
		//assignment, _ := Circuit4VerifyCircuit.GetVerifyCircuitAssignment(Circuit4VerifyCircuit.GetVerifyCircuitParam())
		task := NewTask(i, 2, assignment)
		tasks = append(tasks, task)
	}
	return tasks
}

func NewGenerator(nbTask int) *Generator {
	return &Generator{nbTask: nbTask}
}
