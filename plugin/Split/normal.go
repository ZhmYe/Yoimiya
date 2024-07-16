package Split

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"time"
)

type Groth16NormalRunner struct {
	record   plugin.PluginRecord
	prover   Prover
	verifier Verifier
}

func NewGroth16NormalRunner() Groth16NormalRunner {
	return Groth16NormalRunner{record: plugin.NewPluginRecord()}
}
func (r *Groth16NormalRunner) Prepare(circuit Circuit.TestCircuit) (*cs_bn254.R1CS, witness.Witness) {
	//runtime.GOMAXPROCS(16)
	assignment := circuit.GetAssignment()
	cs, compileTime := circuit.Compile()
	runtime.GC()
	r.record.SetTime("Compile", compileTime)
	//fmt.Println(compileTime)
	setupTime := time.Now()
	pk, vk, err := groth16.Setup(cs)
	if err != nil {
		panic(err)
	}
	runtime.GC()
	//go r.record.MemoryMonitor()
	r.record.SetTime("SetUp", time.Since(setupTime))
	r.prover = NewProver(pk)
	fullWitness, _ := frontend.GenerateWitness(assignment, make([]constraint.ExtraValue, 0), ecc.BN254.ScalarField())
	publicWitness, err := fullWitness.Public()
	if err != nil {
		panic(err)
	}
	r.verifier = NewVerifier(vk, publicWitness)
	//solveTime := time.Now()
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		//commitmentInfo, solution, nbPublic, nbPrivate, err := groth16_bn254.Solve(_r1cs, fullWitness, r.prover.pk)
		//if err != nil {
		//	panic(err)
		//}
		//r.record.SetTime("Solve", time.Since(solveTime))
		return _r1cs, fullWitness
	default:
		panic("Only Support BN254 now!!!")

		//publicWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
	}
	return &cs_bn254.R1CS{}, fullWitness
}
func (r *Groth16NormalRunner) Process(circuit Circuit.TestCircuit) ([]groth16.Proof, error) {
	go r.record.MemoryMonitor()
	Config.Config.CancelSplit()
	proofs := make([]groth16.Proof, 0)
	r1cs, fullWitness := r.Prepare(circuit)
	runtime.GC()

	//publicWitness, _ := frontend.GenerateWitness(assignment, extra, ecc.BN254.ScalarField(), frontend.PublicOnly())
	proveTime := time.Now()
	_, err := r.prover.SolveAndProve(r1cs, fullWitness)
	//var wg sync.WaitGroup
	//wg.Add(2)
	//for i := 0; i < 2; i++ {
	//	go func() {
	//		r.prover.Prove(solution, commitmentInfo, nbPublic, nbPrivate)
	//		wg.Done()
	//	}()
	//}
	//wg.Wait()
	if err != nil {
		panic(err)
	}
	//psroofs = append(proofs, proof)
	r.record.SetTime("Prove", time.Since(proveTime))
	r.record.Finish()
	//for _, proof := range proofs {
	//	isSuccess, err := r.verifier.Verify(proof)
	//	if !isSuccess {
	//		panic(err)
	//	}
	//}
	return proofs, nil
}
func (r *Groth16NormalRunner) Record() {
	r.record.Print()
}
