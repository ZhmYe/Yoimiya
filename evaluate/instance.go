package evaluate

import (
	"Yoimiya/Circuit"
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
	circuit     Circuit.TestCircuit
	memoryAlloc uint64
	//proveMemoryAlloc uint64
	//m    []uint64
	test bool
}

// StartMemoryMonitor 监听内存使用情况
// todo 这里的逻辑
// 目前这里的实现方式是每1s通过runtime.MemStats得到alloc
func (i *Instance) StartMemoryMonitor() {
	startTime := time.Now()
	// 这里可以看到整体内存趋势
	//memorySeq := make([]uint64, 0)
	//var m runtime.MemStats
	//runtime.ReadMemStats(&m)
	//startMemory := m.Alloc
	for {
		if !i.test {
			i.memoryAlloc = uint64(0)
			break
		}
		if time.Since(startTime) >= time.Duration(10)*time.Millisecond {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			nowAlloc := m.Alloc
			if nowAlloc > i.memoryAlloc {
				i.memoryAlloc = nowAlloc
				//i.proveMemoryAlloc = i.memoryAlloc - startMemory
				//fmt.Println(nowAlloc)
			}
			//i.m = append(i.m, nowAlloc)
			//memorySeq = append(memorySeq, nowAlloc)
			startTime = time.Now()
		}
	}
	//fmt.Println(memorySeq)
}
func (i *Instance) GetTotalMemoryAlloc() uint64 {
	//fmt.Println(i.m)
	return i.memoryAlloc
}
func (i *Instance) StartTest() {
	i.test = true
	go func() {
		i.StartMemoryMonitor()
	}()
}

func (i *Instance) TestNormal() Record.Record {
	i.StartTest()
	defer func() {
		i.test = false
	}()
	c := i.circuit
	Config.Config.CancelSplit()
	Record.GlobalRecord.Clear()
	startTime := time.Now()
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	// todo 这里目前只有alone模式加了record
	_, err := split.Split(cs, c.GetAssignment(), 1)
	if err != nil {
		panic(err)
	}
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(int(i.GetTotalMemoryAlloc()))
	//Record.GlobalRecord.SetMemory(int(i.proveMemoryAlloc))
	Record.GlobalRecord.SetSlotTime(time.Since(startTime))
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

// TestNSplit 测试N split带来的内存减少和时间损耗
func (i *Instance) TestNSplit(n int) Record.Record {
	// 这里暂时无视n，n=2
	i.StartTest()
	defer func() {
		i.test = false
	}()
	Config.Config.SwitchToSplit()
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	startTime := time.Now()
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	// todo 这里目前只有alone模式加了record
	_, err := split.Split(cs, c.GetAssignment(), n)
	if err != nil {
		panic(err)
	}
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(int(i.GetTotalMemoryAlloc()))
	//Record.GlobalRecord.SetMemory(int(i.proveMemoryAlloc))
	Record.GlobalRecord.SetSlotTime(time.Since(startTime))
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
	defer func() {
		i.test = false
	}()
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
	defer func() {
		i.test = false
	}()
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	assignmentGenerator := c.GetAssignment
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	//totalProof := make([]split.PackedProof, 0)
	startTime := time.Now()
	for i := 0; i < nbTask; i++ {
		_, err := split.Split(cs, assignmentGenerator(), 1)
		if err != nil {
			panic(err)
		}
		//totalProof = append(totalProof, proofs...)
		runtime.GC()
	}
	Record.GlobalRecord.SetSlotTime(time.Since(startTime))
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(int(i.GetTotalMemoryAlloc()))
	return *Record.GlobalRecord
}
