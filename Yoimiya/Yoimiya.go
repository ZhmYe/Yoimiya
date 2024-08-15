package Yoimiya

import (
	"Yoimiya/Circuit"
)

// Yoimiya 可配置的零知识证明多任务执行框架
// 可以选择三种运行模式：Serial/_Pipeline/SplitPipeline
// 根据split是否为1，可以将前两个运行模式扩展为split/noSplit的串行和pipeline

// YoimiyaRuntimeMode 运行模式
// Serial 串行运行多个任务
// Pipeline 将Solve和Prove进行Pipeline
// SplitPipeline 通过Split将任务数扩大，同时减少每个任务的solve和prove所需的内存从而更好的进行Pipeline比例分配
type YoimiyaRuntimeMode int

const (
	Serial YoimiyaRuntimeMode = iota
	Pipeline
	SplitPipeline
)

type Yoimiya struct {
	mode       YoimiyaRuntimeMode
	nbSplit    int                 // split的个数
	numCPU     int                 // 运行的CPU个数，用于测试不同核数下的性能表现
	nbTask     int                 // 任务数
	solveLimit int                 // solve实例数量
	proveLimit int                 // prove实例数量
	circuit    Circuit.TestCircuit // 任务对应的电路
}

func (y *Yoimiya) SetNumCPU(n int) {
	y.numCPU = n
}
func (y *Yoimiya) SetNbTask(n int) {
	y.nbTask = n
}
func (y *Yoimiya) SetLimit(s int, p int) {
	if s <= 0 {
		s = 1
	}
	if p <= 0 {
		p = 1
	}
	y.solveLimit, y.proveLimit = s, p
}
func (y *Yoimiya) SetNbSplit(s int) {
	y.nbSplit = s
}
func (y *Yoimiya) SetCircuit(option Circuit.CircuitOption) {

}
