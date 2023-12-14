package test

import (
	"S-gnark/backend/groth16"
	"S-gnark/frontend"
	"S-gnark/frontend/cs/r1cs"
	"S-gnark/logger"
	Circuit "S-gnark/statistics/circuit"
	"S-gnark/statistics/utils"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
)

// TestSampleVar test for sample variance of samples
type TestSampleVar struct {
	success  bool          // test pass or not
	prover   Prover        // prover
	verifier Verifier      // verifier
	X        []int         // input X
	Y        int           // input Y
	N        int           // input N
	proof    groth16.Proof // proof
}

func NewTestSampleVar() *TestSampleVar {
	t := new(TestSampleVar)
	t.success, t.N = false, MaxN
	t.X = generateSamples(t.N)
	t.Y = utils.SampleVariance(t.X...)
	return t
}
func (t *TestSampleVar) init() {
	// compiles our circuit into a R1CS
	var circuit Circuit.SampleVarCircuit
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	// groth16 zkSNARK: Setup
	pk, vk, _ := groth16.Setup(ccs)
	// witness definition
	assignment := Circuit.SampleVarCircuit{X1: t.X[0], X2: t.X[1], X3: t.X[2], X4: t.X[3], X5: t.X[4], X6: t.X[5], X7: t.X[6], X8: t.X[7], X9: t.X[8], X10: t.X[9], Y: t.Y, N: t.N}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()
	t.prover = Prover{ccs: ccs, pk: pk, witness: witness}
	t.verifier = Verifier{vk: vk, publicWitness: publicWitness}
}
func (t *TestSampleVar) test() {
	// prover Prove
	t.prover.prove()
	t.proof = t.prover.getProof()
	t.success = t.verifier.verify(t.proof)
}
func (t *TestSampleVar) check() bool {
	return t.success
}
func Test4SampleVar(t *testing.T) {
	log := logger.Logger()
	test := NewTestSampleVar()
	test.init()
	test.test()
	flag := "false"
	if test.check() {
		flag = "true"
	}
	log.Info().Str("Mean Test", flag).Msg("Test LOG")
	//fmt.Println(test.check())
}
