package test

import (
	"S-gnark/backend/groth16"
	"S-gnark/frontend"
	"S-gnark/frontend/cs/r1cs"
	"S-gnark/logger"
	Circuit "S-gnark/statistics/circuit"
	"github.com/consensys/gnark-crypto/ecc"
	"testing"
)

/***
	这里简单做了下层级计算和的尝试
***/

// TestSum test for sum of samples
type TestSum struct {
	success  bool          // test pass or not
	prover   Prover        // prover
	verifier Verifier      // verifier
	X        []int         // input X
	Y        int           // input Y
	proof    groth16.Proof // proof
	sub      []*TestSum    // sub test
}

func NewTestSum(X []int) *TestSum {
	t := new(TestSum)
	t.success = false
	t.X = X
	t.Y = 0
	for _, x := range t.X {
		t.Y += x
	}
	t.sub = make([]*TestSum, 0)
	return t
}
func (t *TestSum) init() {
	flag := false
	inputX := new([MaxN]int)
	// if the number of sample is less than MaxN, then we just use one test to get the result
	if len(t.X) <= MaxN {
		flag = true
		for i := 0; i < MaxN; i++ {
			if i < len(t.X) {
				inputX[i] = t.X[i]
			} else {
				inputX[i] = 0
			}
		}
	} else {
		// if the number of sample is more than MaxN, then we should init several test to get the final result
		instanceNumber := len(t.X)/MaxN + 1 // instance number, sub test number
		// fill the X so that X mod MaxN = 0
		if len(t.X)%MaxN != 0 {
			for i := 0; i < MaxN-len(t.X)%MaxN; i++ {
				t.X = append(t.X, 0)
			}
		}
		// init sub test
		for i := 0; i < instanceNumber; i++ {
			subTest := NewTestSum(t.X[i*MaxN : i*MaxN+MaxN])
			t.sub = append(t.sub, subTest)
		}
		// todo 这里先不写并行了
		// run sub test in parall
		successNumber := 0
		for _, subTest := range t.sub {
			subTest.init()
			subTest.test()
			if subTest.check() {
				successNumber++
			}
		}
		// whether all sub tests pass
		flag = successNumber == len(t.sub)
		// todo 如果flag为false就不用进行下面的test了
		if flag {
			// real input is sub test's Y
			for i := 0; i < MaxN; i++ {
				if i < len(t.sub) {
					inputX[i] = t.sub[i].Y
				} else {
					inputX[i] = 0
				}
			}
		}
	}
	// compiles our circuit into a R1CS
	var circuit Circuit.SumCircuit
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	// groth16 zkSNARK: Setup
	pk, vk, _ := groth16.Setup(ccs)
	// witness definition
	assignment := Circuit.SumCircuit{X1: inputX[0], X2: inputX[1], X3: inputX[2], X4: inputX[3], X5: inputX[4], X6: inputX[5], X7: inputX[6], X8: inputX[7], X9: inputX[8], X10: inputX[9], Y: t.Y}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()
	t.prover = Prover{ccs: ccs, pk: pk, witness: witness}
	t.verifier = Verifier{vk: vk, publicWitness: publicWitness}
}
func (t *TestSum) test() {
	// prover Prove
	t.prover.prove()
	t.proof = t.prover.getProof()
	t.success = t.verifier.verify(t.proof)
}
func (t *TestSum) check() bool {
	return t.success
}
func Test4Sum(t *testing.T) {
	log := logger.Logger()
	test := NewTestSum([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17})
	test.init()
	test.test()
	flag := "false"
	if test.check() {
		flag = "true"
	}
	log.Info().Str("Mean Test", flag).Msg("Test LOG")
	//fmt.Println(test.check())
}
