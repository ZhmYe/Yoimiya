package Parallel

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"sync"
	"time"
)

// ParallelMaster 对于若干个一样的电路，按照最大并发数并行证明，得到Proof
type ParallelMaster struct {
	pk          groth16.ProvingKey
	vk          groth16.VerifyingKey
	cs          constraint.ConstraintSystem
	maxParallel int // 最大并发数
}

func NewParallelMaster(p int) ParallelMaster {
	return ParallelMaster{maxParallel: p}
}
func (m *ParallelMaster) Process(nbTask int, circuit Circuit.TestCircuit) plugin.PluginRecord {
	Config.Config.CancelSplit()
	record := plugin.NewPluginRecord("Parallel")
	cs, compileTime := circuit.Compile()
	m.cs = cs
	record.SetTime("Compile", compileTime)
	setupStartTime := time.Now()
	pk, vk, err := groth16.Setup(m.cs)
	if err != nil {
		panic(err)
	}
	record.SetTime("SetUp", time.Since(setupStartTime))
	m.pk, m.vk = pk, vk
	runtime.GC()
	tasks := make([]witness.Witness, 0)
	for len(tasks) < nbTask {
		assignment := circuit.GetAssignment()
		//tasks = append(tasks, circuit.GetAssignment())
		witness, err := frontend.GenerateWitness(assignment, make([]constraint.ExtraValue, 0), ecc.BN254.ScalarField())
		if err != nil {
			panic(err)
		}
		tasks = append(tasks, witness)
	}
	runtime.GC()
	//proof, err := groth16.Prove(cs, pk.(groth16.ProvingKey), fullWitness)
	startTime := time.Now()
	nbIter := nbTask / m.maxParallel
	go record.MemoryMonitor()
	for i := 0; i < nbIter; i++ {
		var wg sync.WaitGroup
		wg.Add(m.maxParallel)
		//startTime := time.Now()
		for j := 0; j < m.maxParallel; j++ {
			tmp := j
			go func(index int) {
				_, err := groth16.Prove(cs, pk.(groth16.ProvingKey), tasks[index])
				wg.Done()
				if err != nil {
					panic(err)
				}

			}(tmp)
		}
		wg.Wait()
		//fmt.Println(time.Since(startTime))
		runtime.GC()
		//_, err := groth16.Prove(cs, pk.(groth16.ProvingKey), fullWitness)
	}
	if nbTask%m.maxParallel != 0 {
		var wg sync.WaitGroup
		wg.Add(nbTask % m.maxParallel)
		for j := nbIter * m.maxParallel; j < nbTask; j++ {
			tmp := j
			go func(index int) {
				_, err := groth16.Prove(cs, pk.(groth16.ProvingKey), tasks[index])
				wg.Done()
				if err != nil {
					panic(err)
				}

			}(tmp)
		}
		wg.Wait()
	}
	proveTime := time.Since(startTime)
	record.SetTime("Prove", proveTime)
	record.Finish()
	return record
}
