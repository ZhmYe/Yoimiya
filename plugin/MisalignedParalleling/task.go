package MisalignedParalleling

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"time"
)

type Task struct {
	receipt TaskReceipt         // 回执，记录运行情况
	cut     int                 // 一共多少个电路
	phase   int                 // 目前进行到第几个电路
	proofs  []split.PackedProof // 最后得到的proof
	pli     frontend.PackedLeafInfo
	//assignment  frontend.Circuit        // 输入
	extra       []constraint.ExtraValue // 所有的extra
	id          int                     // Task id,用于debug
	coordinator *Coordinator            // 任务的协调者
	finish      bool
	startTime   time.Time
	//block       bool
}

func NewTask(id int, assignment frontend.Circuit, cut int) *Task {
	return &Task{
		receipt:   NewTaskReceipt(),
		cut:       cut,
		phase:     0,
		proofs:    make([]split.PackedProof, 0),
		pli:       frontend.GetPackedLeafInfoFromAssignment(assignment),
		extra:     make([]constraint.ExtraValue, 0),
		id:        id,
		finish:    false,
		startTime: time.Now(),
		//block:   false,
	}
}
func (t *Task) SetCoordinator(c *Coordinator) {
	t.coordinator = c
}
func (t *Task) Request() {
	if t.phase == t.cut {
		t.finish = true
		t.receipt.TotalTime(time.Since(t.startTime))
		return
	}
	t.coordinator.HandleRequest(t.id, t.pli, t.extra, t.phase)
}
func (t *Task) HandleResponse(receipt SlotReceipt) {
	if t.id != receipt.tID || t.phase != receipt.sID {
		return
	}
	t.receipt.UpdateProveTime(receipt.proveTime)
	t.receipt.SetWaitTime(time.Since(t.startTime) - receipt.proveTime)
	t.extra = append(t.extra, receipt.extra...)
	// todo receipt.proof 这里暂时先不管能不能verify过
	//packedProof := receipt.proof
	//err := groth16.Verify(packedProof.GetProof(), packedProof.GetVerifyingKey(), packedProof.GetPublicWitness())
	//if err != nil {
	//	panic(err)
	//}
	t.phase++
	t.Request()
	//t.block = false
}

//func (t *Task) Run() {
//	t.startTime = time.Now()
//	for {
//		if t.finish {
//			break
//		}
//		if !t.block {
//			t.Request()
//			t.block = true
//		}
//	}
//	t.receipt.TotalTime(time.Since(t.startTime))
//}

//func (t *Task) GenerateWitness()
