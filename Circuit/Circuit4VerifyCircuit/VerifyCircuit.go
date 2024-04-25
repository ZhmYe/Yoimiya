package Circuit4VerifyCircuit

import (
	groth16backend_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"Yoimiya/std/algebra/emulated/sw_bn254"
	"Yoimiya/std/math/emulated"
	"Yoimiya/std/math/emulated/emparams"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	fr_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"time"
)

type VerifyCircuit struct {
	assignmentGenerator func() frontend.Circuit // 生成assignment
	outerCircuit        frontend.Circuit
}

func NewVerifyCircuit() VerifyCircuit {
	c := VerifyCircuit{
		assignmentGenerator: nil,
		outerCircuit:        nil,
	}
	c.Init()
	return c
}

// GetAssignment 获取测试电路的随机Assignment
func (c *VerifyCircuit) GetAssignment() frontend.Circuit {
	return c.assignmentGenerator()
}

// Init 初始化
func (c *VerifyCircuit) Init() {
	innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray := GetVerifyCircuitParam()
	var circuitVk [LENGTH]VerifyingKey[sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl]

	for j := range innerCcsArray {
		innerVK := innerVKArray[j]

		tVk, _ := innerVK.(*groth16backend_bn254.VerifyingKey)
		e, err := bn254.Pair([]bn254.G1Affine{tVk.G1.Alpha}, []bn254.G2Affine{tVk.G2.Beta}) // compute E
		circuitVk[j].E = sw_bn254.NewGTEl(e)
		circuitVk[j].G1.K = make([]sw_bn254.G1Affine, len(tVk.G1.K))
		for i := range circuitVk[j].G1.K {
			circuitVk[j].G1.K[i] = sw_bn254.NewG1Affine(tVk.G1.K[i])
		}
		var deltaNeg, gammaNeg bn254.G2Affine
		deltaNeg.Neg(&tVk.G2.Delta)
		gammaNeg.Neg(&tVk.G2.Gamma)
		circuitVk[j].G2.DeltaNeg = sw_bn254.NewG2Affine(deltaNeg)
		circuitVk[j].G2.GammaNeg = sw_bn254.NewG2Affine(gammaNeg)
		if err != nil {
			panic(err)
		}
	}

	// ScalarField is the [emulated.FieldParams] impelementation of the curve scalar field.
	type ScalarField = emulated.BN254Fr

	// get outer circuit witness
	var circuitWitness [LENGTH]Witness[ScalarField]
	for j, innerWitness := range innerWitnessArray {
		pubw, _ := innerWitness.Public()
		vec := pubw.Vector()
		vect, _ := vec.(fr_bn254.Vector)
		for i := range vect {
			circuitWitness[j].Public = append(circuitWitness[j].Public, emulated.ValueOf[emparams.BN254Fr](vect[i]))
		}
	}

	// get outer circuit proof
	var circuitProof [LENGTH]Proof[sw_bn254.G1Affine, sw_bn254.G2Affine]
	for i, innerProof := range innerProofArray {
		tpoof, _ := innerProof.(*groth16backend_bn254.Proof)
		circuitProof[i].Ar = sw_bn254.NewG1Affine(tpoof.Ar)
		circuitProof[i].Krs = sw_bn254.NewG1Affine(tpoof.Krs)
		circuitProof[i].Bs = sw_bn254.NewG2Affine(tpoof.Bs)
	}

	//fmt.Println("circuitVk: ", circuitVk)
	//fmt.Println("circuitWitness: ", circuitWitness)
	//fmt.Println("circuitProof: ", circuitProof)

	c.outerCircuit = &OuterCircuit[ScalarField, sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl]{
		InnerWitness: func(innerCcsArray [LENGTH]constraint.ConstraintSystem) [LENGTH]Witness[ScalarField] {
			witnessArray := [LENGTH]Witness[ScalarField]{}
			for i, innerCcs := range innerCcsArray {
				witnessArray[i].Public = make([]emulated.Element[ScalarField], innerCcs.GetNbPublicVariables()-1)
			}
			return witnessArray
		}(innerCcsArray),
		VerifyingKey: func(innerCcsArray [LENGTH]constraint.ConstraintSystem) [LENGTH]VerifyingKey[sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl] {
			verifyingKeyArray := [LENGTH]VerifyingKey[sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl]{}
			for i, innerCcs := range innerCcsArray {
				verifyingKeyArray[i].G1.K = make([]sw_bn254.G1Affine, innerCcs.GetNbPublicVariables())
			}
			return verifyingKeyArray
		}(innerCcsArray),
	}
	c.assignmentGenerator = func() frontend.Circuit {
		return &OuterCircuit[ScalarField, sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl]{
			InnerWitness: circuitWitness,
			Proof:        circuitProof,
			VerifyingKey: circuitVk,
		}
	}
}

func (c *VerifyCircuit) Compile() (constraint.ConstraintSystem, time.Duration) {
	startTime := time.Now()
	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, c.outerCircuit)
	if err != nil {
		panic(err)
	}
	compileTime := time.Since(startTime)
	return outerCcs, compileTime
}
