package plugin

import "time"

type TaskReceipt struct {
	valid     bool
	wait_time time.Duration // 当前任务从创建到被处理的时间
	//set_up_time time.Duration // 当前任务set up阶段花费的总时间
	prove_time time.Duration // 当前任务prove阶段花费的时间
	block_time time.Duration // 当前任务除去set up, prove以外剩下的阻塞时间
	total_time time.Duration // 任务从创建到结束的总时间, 所有任务中这个时间最长的就是最后的总时间(假设所有任务同时到达)
	verify     bool          // verify结果
}

// 设置一些可以直接得到的时间

func (r *TaskReceipt) SetWaitTime(waitTime time.Duration) {
	if r.valid {
		return
	}
	r.valid = true
	r.wait_time = waitTime
}

//	func (r *TaskReceipt) Time4SetUp(setupTime time.Duration) {
//		r.set_up_time = setupTime
//	}

func (r *TaskReceipt) UpdateProveTime(proveTime time.Duration) {
	r.prove_time += proveTime
}
func (r *TaskReceipt) TotalTime(total time.Duration) {
	r.total_time = total
}

// ComputeBlockTime block的时间需要额外计算
func (r *TaskReceipt) ComputeBlockTime() {
	r.block_time = r.total_time - r.wait_time - r.prove_time
}
func (r *TaskReceipt) SetVerifyFlag(flag bool) {
	r.verify = flag
}
func NewTaskReceipt() TaskReceipt {
	return TaskReceipt{
		wait_time: 0,
		//set_up_time: 0,
		prove_time: 0,
		block_time: 0,
		total_time: 0,
		verify:     true,
		valid:      false,
	}
}
