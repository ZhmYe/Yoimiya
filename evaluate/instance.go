package evaluate

import (
	"Yoimiya/Config"
	"Yoimiya/MisalignedParalleling"
	"Yoimiya/Record"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	"Yoimiya/frontend/split"
	"fmt"
	"time"
)

// Instance 测试实例
type Instance struct {
	circuit testCircuit
}

// TestNSplit 测试N split带来的内存减少和时间损耗
func (i *Instance) TestNSplit(n int) Record.Record {
	// 这里暂时无视n，n=2
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	startTime := time.Now()
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	// todo 这里目前只有alone模式加了record
	proofs, err := split.Split(cs, c.GetAssignment(), split.NewParam(true, Config.Config.IsCluster(), n, false))
	if err != nil {
		panic(err)
	}
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(0)

	fmt.Println("Split Circuit Time:", time.Since(startTime))
	for i, packedProof := range proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		err := groth16.Verify(proof, verifyKey, publicWitness)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Proof ", i, " Verify Success...")
		}
	}
	return *Record.GlobalRecord
}

// TestMisalignedParalleling 测试错位并行的总运行时间和内存
func (i *Instance) TestMisalignedParalleling(nbTask int, cut int) Record.Record {
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
	Record.GlobalRecord.SetMemory(0)
	return *Record.GlobalRecord
}

// TestSerialRunning 测试串行运行的总时间和内存使用
func (i *Instance) TestSerialRunning(nbTask int) Record.Record {
	c := i.circuit
	Record.GlobalRecord.Clear() // 清除record
	assignmentGenerator := c.GetAssignment
	cs, compileTime := c.Compile()
	Record.GlobalRecord.SetCompileTime(compileTime) // 记录compile时间
	totalProof := make([]split.PackedProof, 0)
	startTime := time.Now()
	for i := 0; i < nbTask; i++ {
		proofs, err := split.Split(cs, assignmentGenerator(), split.NewParam(false, false, -1, false))
		if err != nil {
			panic(err)
		}
		totalProof = append(totalProof, proofs...)
	}
	Record.GlobalRecord.SetSlotTime(time.Since(startTime))
	// todo 这里加上内存的测试逻辑
	Record.GlobalRecord.SetMemory(0)
	return *Record.GlobalRecord
}
