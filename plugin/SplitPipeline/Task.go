package SplitPipeline

import (
	"Yoimiya/Circuit"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"Yoimiya/plugin"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"sync"
	"time"
)

// Task 将pipeline表示为若干个任务
// 会有一个TaskPool，每次从TaskPool中取出一个进行solve和prove
type Task struct {
	tID    int                     // 这里的id是真实的一个任务的id，一个任务会有s个子任务，这s个子任务的tID相同
	phase  int                     // 上述子任务对应的阶段，也就是子任务的id
	extra  []constraint.ExtraValue // 额外的public input
	pli    frontend.PackedLeafInfo // 原有的assignment
	count  int                     // split数量，也就是phase的最大值
	proofs []split.PackedProof     // 该任务的所有proof
	finish bool
	//wg     *sync.WaitGroup         // 用来说明所有任务全部完成
	execLock sync.Mutex
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

func (t *Task) UpdateExtra(extra []constraint.ExtraValue) {
	t.extra = append(t.extra, extra...)
}

func (t *Task) Process(pk groth16.ProvingKey, ccs constraint.ConstraintSystem, inputID []int, vk groth16.VerifyingKey) {
	if t.finish {
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	//fmt.Println(inputID)
	//t.execLock.Lock()
	witness, err := frontend.GenerateSplitWitnessFromPli(t.pli, inputID, t.extra, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)
	}
	prover := plugin.NewProver(pk)
	startTime := time.Now()
	proof, err := prover.SolveAndProve(ccs.(*cs_bn254.R1CS), witness)
	//t.execLock.Unlock()
	fmt.Printf("solve and prove: %v, \n", time.Since(startTime))
	//proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		panic(err)
	}
	publicWitness, err := witness.Public()
	if err != nil {
		panic(err)
	}
	t.proofs = append(t.proofs, split.NewPackedProof(proof, vk, publicWitness))
}

func (t *Task) SyncProcess(pk groth16.ProvingKey, ccs constraint.ConstraintSystem, inputID []int, vk groth16.VerifyingKey, solveLock chan int, ProveLock *sync.Mutex, nbCommit *int) {
	//if t.phase == t.count-1 {
	//	return
	//}
	if t.finish {
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	//t.execLock.Lock()

	witness, err := frontend.GenerateSplitWitnessFromPli(t.pli, inputID, t.extra, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)
	}
	prover := plugin.NewProver(pk)
	solveLock <- 1
	startTime := time.Now()
	commitmentsInfo, solution, nbPublic, nbPrivate := prover.Solve(ccs.(*cs_bn254.R1CS), witness)
	fmt.Printf("%d solveTime: %s\n", t.tID, time.Since(startTime))
	<-solveLock
	//time.Sleep(10 * time.Minute)
	newExtra := split.GetExtra(ccs)
	t.UpdateExtra(newExtra)

	//t.execLock.Unlock()

	//mutex.Lock()
	//*channel <- 1
	//<-*solveLock
	//ProveLock.Lock()

	time.Sleep(100 * time.Second)

	ProveLock.Lock()
	startTimeForOneProcess := time.Now()
	proof, err := prover.Prove(*solution, commitmentsInfo, nbPublic, nbPrivate)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d ProveTime: %s\n", t.tID, time.Since(startTimeForOneProcess))
	publicWitness, err := witness.Public()
	if err != nil {
		panic(err)
	}
	t.proofs = append(t.proofs, split.NewPackedProof(proof, vk, publicWitness))
	//runtime.GC()
	ProveLock.Unlock()
	//<-*channel
	if !t.Next() {
		//wg.Done()
		*nbCommit++
		//fmt.Println(*nbCommit)
	}

	/*go func() {
		ProveLock.Lock()
		startTimeForOneProcess := time.Now()
		proof, err := prover.Prove(*solution, commitmentsInfo, nbPublic, nbPrivate)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d ProveTime: %s\n", t.tID, time.Since(startTimeForOneProcess))
		publicWitness, err := witness.Public()
		if err != nil {
			panic(err)
		}
		t.proofs = append(t.proofs, split.NewPackedProof(proof, vk, publicWitness))
		//runtime.GC()
		ProveLock.Unlock()
		//<-*channel
		if !t.Next() {
			//wg.Done()
			*nbCommit++
			//fmt.Println(*nbCommit)
		}

	}()*/
}
