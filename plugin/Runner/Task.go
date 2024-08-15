package Runner

import (
	"Yoimiya/Circuit"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
)

// Task 这里的Task用于被Runner调度，本身不具备运行能力，只负责传递参数
type Task struct {
	tID    int                     // 这里的id是真实的一个任务的id，一个任务会有s个子任务，这s个子任务的tID相同
	phase  int                     // 上述子任务对应的阶段，也就是子任务的id
	extra  []constraint.ExtraValue // 额外的public input
	pli    frontend.PackedLeafInfo // 原有的assignment
	count  int                     // split数量，也就是phase的最大值
	proofs []split.PackedProof     // 该任务的所有proof
	finish bool
	//wg     *sync.WaitGroup         // 用来说明所有任务全部完成
	//execLock sync.Mutex
}
type Request struct {
	Pli   frontend.PackedLeafInfo
	Extra []constraint.ExtraValue
	ID    int
	Phase int
}
type Response struct {
	Proof split.PackedProof
	Extra []constraint.ExtraValue
	ID    int
	Phase int
}

func NewTask(circuit Circuit.TestCircuit, split int, id int) *Task {
	assignment := circuit.GetAssignment()
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	return &Task{
		tID:    id,
		phase:  0,
		extra:  make([]constraint.ExtraValue, 0),
		count:  split,
		pli:    pli,
		finish: false,
	}
}
func (t *Task) Next() bool {
	if t.phase == t.count-1 {
		//t.wg.Done() // 该任务结束
		//fmt.Println(len(t.proofs))
		t.finish = true
		return false
	}
	t.phase++
	return true
}
func (t *Task) Params() Request {
	return Request{
		Pli:   t.pli,
		Extra: t.extra,
		ID:    t.tID,
	}
}
func (t *Task) UpdateExtra(extra []constraint.ExtraValue) {
	t.extra = append(t.extra, extra...)
}
func (t *Task) Done() bool {
	return t.finish
}

func (t *Task) HandleResponse(r Response) bool {
	proof, newExtra, id := r.Proof, r.Extra, r.ID
	if id != t.tID {
		panic("Error Response Task ID")
	}
	t.proofs = append(t.proofs, proof)
	t.UpdateExtra(newExtra)
	return t.Next()
}
func (t *Task) ID() int {
	return t.tID
}
