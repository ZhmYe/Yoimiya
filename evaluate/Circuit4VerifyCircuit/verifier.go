package Circuit4VerifyCircuit

import (
	"S-gnark/backend/groth16"
	groth16backend_bn254 "S-gnark/backend/groth16/bn254"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	"S-gnark/frontend"
	"S-gnark/frontend/cs/r1cs"
	"S-gnark/std/algebra"
	"S-gnark/std/algebra/emulated/sw_bn254"
	"S-gnark/std/math/emulated"
	"S-gnark/std/math/emulated/emparams"
	"S-gnark/test"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	fr_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Proof is a typed Groth16 proof of SNARK. Use [ValueOfProof] to initialize the
// witness from the native proof.
type Proof[G1El algebra.G1ElementT, G2El algebra.G2ElementT] struct {
	Ar, Krs G1El
	Bs      G2El
}

// VerifyingKey is a typed Groth16 verifying key for checking SNARK proofs. For
// witness creation use the method [ValueOfVerifyingKey] and for stub
// placeholder use [PlaceholderVerifyingKey].
type VerifyingKey[G1El algebra.G1ElementT, G2El algebra.G2ElementT, GtEl algebra.GtElementT] struct {
	E  GtEl
	G1 struct{ K []G1El }
	G2 struct{ GammaNeg, DeltaNeg G2El }
}

// Witness is a public witness to verify the SNARK proof against. For assigning
// witness use [ValueOfWitness] and to create stub witness for compiling use
// [PlaceholderWitness].
type Witness[FR emulated.FieldParams] struct {
	// Public is the public inputs. The first element does not need to be one
	// wire and is added implicitly during verification.
	Public []emulated.Element[FR]
}

// Verifier verifies Groth16 proofs.
type Verifier[FR emulated.FieldParams, G1El algebra.G1ElementT, G2El algebra.G2ElementT, GtEl algebra.GtElementT] struct {
	curve   algebra.Curve[FR, G1El]
	pairing algebra.Pairing[G1El, G2El, GtEl]
}

// NewVerifier returns a new [Verifier] instance using the curve and pairing
// interfaces. Use methods [algebra.GetCurve] and [algebra.GetPairing] to
// initialize the instances.
func NewVerifier[FR emulated.FieldParams, G1El algebra.G1ElementT, G2El algebra.G2ElementT, GtEl algebra.GtElementT](curve algebra.Curve[FR, G1El], pairing algebra.Pairing[G1El, G2El, GtEl]) *Verifier[FR, G1El, G2El, GtEl] {
	return &Verifier[FR, G1El, G2El, GtEl]{
		curve:   curve,
		pairing: pairing,
	}
}

// AssertProof asserts that the SNARK proof holds for the given witness and
// verifying key.
func (v *Verifier[FR, G1El, G2El, GtEl]) AssertProof(vk VerifyingKey[G1El, G2El, GtEl], proof Proof[G1El, G2El], witness Witness[FR]) error {
	inP := make([]*G1El, len(vk.G1.K)-1) // first is for the one wire, we add it manually after MSM
	for i := range inP {
		inP[i] = &vk.G1.K[i+1]
	}
	inS := make([]*emulated.Element[FR], len(witness.Public))
	for i := range inS {
		inS[i] = &witness.Public[i]
	}
	kSum, err := v.curve.MultiScalarMul(inP, inS)
	if err != nil {
		return fmt.Errorf("multi scalar mul: %w", err)
	}
	kSum = v.curve.Add(kSum, &vk.G1.K[0])
	pairing, err := v.pairing.Pair([]*G1El{kSum, &proof.Krs, &proof.Ar}, []*G2El{&vk.G2.GammaNeg, &vk.G2.DeltaNeg, &proof.Bs})
	if err != nil {
		return fmt.Errorf("pairing: %w", err)
	}
	v.pairing.AssertIsEqual(pairing, &vk.E)
	return nil
}

func getCircuitVkWitnessPublic(
	assert *test.Assert,
	innerCcsArray [LENGTH]constraint.ConstraintSystem,
	innerVKArray [LENGTH]groth16.VerifyingKey,
	innerWitnessArray [LENGTH]witness.Witness,
	innerProofArray [LENGTH]groth16.Proof) (
	constraint.ConstraintSystem,
	groth16.ProvingKey,
	groth16.VerifyingKey,
	witness.Witness,
	witness.Witness) {

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
		assert.NoError(err)
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

	outerCircuit := &OuterCircuit[ScalarField, sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl]{
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

	outerAssignment := &OuterCircuit[ScalarField, sw_bn254.G1Affine, sw_bn254.G2Affine, sw_bn254.GTEl]{
		InnerWitness: circuitWitness,
		Proof:        circuitProof,
		VerifyingKey: circuitVk,
	}

	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, outerCircuit)
	assert.NoError(err)
	outerPK, outerVK, err := groth16.Setup(outerCcs)

	full, err := frontend.NewWitness(outerAssignment, ecc.BN254.ScalarField())
	public, err := frontend.NewWitness(outerAssignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	return outerCcs, outerPK, outerVK, full, public
}
