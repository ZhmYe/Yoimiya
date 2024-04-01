package test

import (
	"S-gnark/Config"
	"S-gnark/backend/groth16"
	"S-gnark/frontend"
	"S-gnark/frontend/cs/r1cs"
	"S-gnark/frontend/split"
	"S-gnark/logger"
	Circuit "S-gnark/statistics/circuit"
	"S-gnark/statistics/utils"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
	"time"
)

// TestMean test for mean of samples
type TestMean struct {
	success  bool                     // test pass or not
	prover   Prover                   // prover
	verifier Verifier                 // verifier
	X        [10000]frontend.Variable // input X
	Y        int                      // input Y
	N        int                      // input N
	proof    groth16.Proof            // proof
}

func NewTestMean() *TestMean {
	t := new(TestMean)
	t.success, t.N = true, 10000
	SampleVariable, SampleInt := generateSamples(t.N)
	t.X = SampleVariable
	t.Y = utils.Mean(SampleInt...)
	return t
}
func (t *TestMean) init() {
	// compiles our circuit into a R1CS
	var circuit Circuit.MeanCircuit
	assignment := Circuit.MeanCircuit{X: t.X, Y: t.Y, N: t.N}
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	// groth16 zkSNARK: Setup
	startTime := time.Now()
	proofs, err := split.Split(ccs, &assignment, split.NewParam(true, Config.Config.IsCluster(), 2, false))
	if err != nil {
		return
	}
	fmt.Println(len(proofs))
	fmt.Println("Split Circuit Time:", time.Since(startTime))
	for _, packedProof := range proofs {
		proof := packedProof.GetProof()
		verifyKey := packedProof.GetVerifyingKey()
		publicWitness := packedProof.GetPublicWitness()
		verifier := Verifier{
			vk:            verifyKey,
			publicWitness: publicWitness,
		}
		if !verifier.verify(proof) {
			t.success = false
		}
	}
	//for _, split := range splits {
	//	switch _split := split.(type) {
	//	case *cs_bn254.R1CS:
	//		fmt.Println(_split.Sit.GetTotalInstructionNumber())
	//	default:
	//		panic("Only Support bn254 r1cs now...")
	//	}
	//}
	//pk, vk, _ := groth16.Setup(ccs)
	//fmt.Println("Set Up Time:", time.Since(startTime))
	// witness definition
	//witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	//publicWitness, _ := witness.Public()
	//t.prover = Prover{ccs: ccs, pk: pk, witness: witness}
	//t.verifier = Verifier{vk: vk, publicWitness: publicWitness}
}
func (t *TestMean) test() {
	// prover Prove
	t.prover.prove()
	t.proof = t.prover.getProof()
	t.success = t.verifier.verify(t.proof)
}
func (t *TestMean) check() bool {
	return t.success
}
func Test4Mean(t *testing.T) {
	log := logger.Logger()
	test := NewTestMean()
	test.init()
	//test.test()
	flag := "false"
	if test.check() {
		flag = "true"
	}
	log.Info().Str("Mean Test", flag).Msg("Test LOG")
	//fmt.Println(test.check())
}
