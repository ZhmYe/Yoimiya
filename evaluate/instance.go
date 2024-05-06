package evaluate

import (
	"Yoimiya/Config"
	"Yoimiya/MisalignedParalleling"
	"Yoimiya/Record"
	"Yoimiya/constraint"
	"Yoimiya/frontend/split"
	"fmt"
	"runtime"
	"time"
)

// Instance 测试实例
type Instance struct {
	circuit     testCircuit
	memoryAlloc uint64
	test        bool
}

// StartMemoryMonitor 监听内存使用情况
// todo 这里的逻辑
func (i *Instance) StartMemoryMonitor() {
	startTime := time.Now()
	for {
		if !i.test {
			i.memoryAlloc = uint64(0)
			break
		}
		if time.Since(startTime) >= time.Duration(1)*time.Second {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			nowAlloc := m.Alloc
			if nowAlloc > i.memoryAlloc {
				i.memoryAlloc = nowAlloc
				fmt.Println(nowAlloc)
			}
		}

	}
}
func (i *Instance) GetTotalMemoryAlloc() uint64 {
	return i.memoryAlloc
}
func (i *Instance) StartTest() {
	i.test = true
	go func() {
		//i.StartMemoryMonitor()
	}()
	//defer func() {
	//	i.test = false
	//}()
}

// TestNSplit 测试N split带来的内存减少和时间损耗
func (i *Instance) TestNSplit(n int) Record.Record {
	// 这里暂时无视n，n=2
	i.StartTest()
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	startTime := time.Now()
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	// todo 这里目前只有alone模式加了record
	_, err := split.Split(cs, c.GetAssignment(), split.NewParam(true, Config.Config.IsCluster(), n, false))
	if err != nil {
		panic(err)
	}
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(int(i.GetTotalMemoryAlloc()))

	fmt.Println("Split Circuit Time:", time.Since(startTime))
	//for i, packedProof := range proofs {
	//	proof := packedProof.GetProof()
	//	verifyKey := packedProof.GetVerifyingKey()
	//	publicWitness := packedProof.GetPublicWitness()
	//	err := groth16.Verify(proof, verifyKey, publicWitness)
	//	if err != nil {
	//		fmt.Println(err)
	//	} else {
	//		fmt.Println("Proof ", i, " Verify Success...")
	//	}
	//}
	return *Record.GlobalRecord
}

// TestMisalignedParalleling 测试错位并行的总运行时间和内存
func (i *Instance) TestMisalignedParalleling(nbTask int, cut int) Record.Record {
	i.StartTest()
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	master := MisalignedParalleling.NewMParallelingMaster()
	assignmentGenerator := c.GetAssignment
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	csGenerator := func() constraint.ConstraintSystem {
		return cs
	}
	master.Initialize(nbTask, cut, csGenerator, assignmentGenerator)
	master.Start()
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(int(i.GetTotalMemoryAlloc()))
	return *Record.GlobalRecord
}

// TestSerialRunning 测试串行运行的总时间和内存使用
func (i *Instance) TestSerialRunning(nbTask int) Record.Record {
	i.StartTest()
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	assignmentGenerator := c.GetAssignment
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	//totalProof := make([]split.PackedProof, 0)
	startTime := time.Now()
	for i := 0; i < nbTask; i++ {
		_, err := split.Split(cs, assignmentGenerator(), split.NewParam(false, false, -1, false))
		if err != nil {
			panic(err)
		}
		//totalProof = append(totalProof, proofs...)
	}
	Record.GlobalRecord.SetSlotTime(time.Since(startTime))
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(int(i.GetTotalMemoryAlloc()))
	return *Record.GlobalRecord
}
