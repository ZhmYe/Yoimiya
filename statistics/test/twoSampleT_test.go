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
type TestTwoSampleT struct {
	success  bool          // test pass or not
	prover   Prover        // prover
	verifier Verifier      // verifier
	X        []int         // input X
	Y        []int         // input Y
	m1       int           // input m1
	m2       int           //input m2
	T        int           // staitc t
	proof    groth16.Proof // proof
}

func NewTestTwoSampleT() *TestTwoSampleT {
	t := new(TestTwoSampleT)
	t.success = false
	t.m1, t.m2 = MaxN, MaxN
	_, t.X = generateSamples(t.m1)
	_, t.Y = generateSamples(t.m2)
	t.T = utils.TwoSampleT(t.X, t.Y)
	return t
}
func (t *TestTwoSampleT) init() {
	// compiles our circuit into a R1CS
	var circuit Circuit.TwoSampleTCircuit
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	// groth16 zkSNARK: Setup
	pk, vk, _ := groth16.Setup(ccs)
	assignment := Circuit.TwoSampleTCircuit{X1: t.X[0], X2: t.X[1], X3: t.X[2], X4: t.X[3], X5: t.X[4], X6: t.X[5], X7: t.X[6], X8: t.X[7], X9: t.X[8], X10: t.X[9],
		Y1: t.Y[0], Y2: t.Y[1], Y3: t.Y[2], Y4: t.Y[3], Y5: t.Y[4], Y6: t.Y[5], Y7: t.Y[6], Y8: t.Y[7], Y9: t.Y[8], Y10: t.Y[9],
		M1: t.m1, M2: t.m2, T: t.T}
	//assignment := Circuit.TwoSampleTCircuit{X1: t.X[0], X2: t.X[1], X3: t.X[2], X4: t.X[3], X5: t.X[4], X6: t.X[5], X7: t.X[6], X8: t.X[7], X9: t.X[8], X10: t.X[9], N: t.N}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()
	t.prover = Prover{ccs: ccs, pk: pk, witness: witness}
	t.verifier = Verifier{vk: vk, publicWitness: publicWitness}
}
func (t *TestTwoSampleT) test() {
	// prover Prove
	t.prover.prove()
	t.proof = t.prover.getProof()
	t.success = t.verifier.verify(t.proof)
}
func (t *TestTwoSampleT) check() bool {
	return t.success
}
func Test4TwoSampleT(t *testing.T) {
	log := logger.Logger()
	test := NewTestTwoSampleT()
	test.init()
	test.test()
	flag := "false"
	if test.check() {
		flag = "true"
	}
	log.Info().Str("Mean Test", flag).Msg("Test LOG")
	//fmt.Println(test.check())
}