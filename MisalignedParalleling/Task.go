package MisalignedParalleling

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"fmt"
	"sync"
	"time"
)

// Task 用于表示每个并行的基于相同电路的（输入不同）的任务
type Task struct {
	processTime time.Time               // 该任务开始被处理的时间
	startTime   time.Time               // 该任务的第一个电路开始运行的时间
	endTime     time.Time               // 该任务所有电路运行完成的时间
	cut         int                     // 一共多少个电路
	phase       int                     // 目前进行到第几个电路
	proofs      []split.PackedProof     // 最后得到的proof
	assignment  frontend.Circuit        // 输入
	extra       []constraint.ExtraValue // 所有的extra
	mutex       sync.Mutex
	id          int // Task id,用于debug
}
type Tasks []*Task

func (t *Tasks) SetProcessTime() {
	for i, task := range *t {
		task.SetProcessTime()
		(*t)[i] = task
	}
}
func NewTask(id int, cut int, assignment frontend.Circuit) *Task {
	task := &Task{
		id:         id,
		cut:        cut,
		assignment: assignment,
		extra:      make([]constraint.ExtraValue, 0),
		proofs:     make([]split.PackedProof, 0),
		phase:      -1,
	}
	//task.SetProcessTime()
	return task
}

// Process 执行第i个电路，得到证明并更新extra
func (t *Task) Process(phase int, cs PackedConstraintSystem) {
	t.mutex.Lock()
	if phase != t.phase+1 {
		panic("Has Some Phase Not Process!!!")
	}
	defer t.Advance()
	if t.phase == -1 {
		// 执行第一个电路
		t.SetStartTime()
	}
	t.AppendProof(cs.Process(t.extra, t.assignment))
	forwardOutput := cs.GetForwardOutput()
	t.UpdateExtra(forwardOutput)
}

// Verify 验证所有proof
func (t *Task) Verify() bool {
	if len(t.proofs) != t.cut {
		panic("Proofs Not Complete!!!")
	}
	for _, packedProof := range t.proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		err := groth16.Verify(proof, verifyKey, publicWitness)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	return true
}

// SetProcessTime 在该任务被创建的时候就调用
func (t *Task) SetProcessTime() {
	t.processTime = time.Now()
}

// SetStartTime 在该任务的第一个电路开始运行的时候调用
func (t *Task) SetStartTime() {
	t.startTime = time.Now()
}

// SetEndTime 在该任务所有电路运行完后调用
func (t *Task) SetEndTime() {
	t.endTime = time.Now()
}

// WaitTime 返回任务真正开始运行的等待时间
func (t *Task) WaitTime() time.Duration {
	return t.startTime.Sub(t.processTime)
}

// RunTime 返回任务运行时间
func (t *Task) RunTime() time.Duration {
	return t.endTime.Sub(t.startTime)
}

// Latency 任务总延时
func (t *Task) Latency() time.Duration {
	return t.RunTime() + t.WaitTime()
}
func (t *Task) Advance() {
	t.phase++
	if t.phase == t.cut {
		// 已经完成了所有电路
		t.SetEndTime()
	}
	t.mutex.Unlock()
}
func (t *Task) AppendProof(proof split.PackedProof) {
	//t.proofs = append(t.proofs, proof)
}

func (t *Task) UpdateExtra(forwardOutput []constraint.ExtraValue) {
	t.extra = append(t.extra, forwardOutput...)
}
