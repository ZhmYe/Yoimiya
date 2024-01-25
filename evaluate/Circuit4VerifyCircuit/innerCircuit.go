package Circuit4VerifyCircuit

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	"S-gnark/frontend"
	"S-gnark/frontend/cs/r1cs"
	"S-gnark/test"
	"math/big"
)

type InnerCircuit struct {
	P, Q frontend.Variable
	N    frontend.Variable `gnark:",public"`
}

func (c *InnerCircuit) Define(api frontend.API) error {
	res := api.Mul(c.P, c.Q)
	api.AssertIsEqual(res, c.N)
	return nil
}

func getInner(assert *test.Assert, field *big.Int) (constraint.ConstraintSystem, groth16.VerifyingKey, witness.Witness, groth16.Proof) {
	// compiles our circuit into a R1CS
	innerCcs, err := frontend.Compile(field, r1cs.NewBuilder, &InnerCircuit{})
	assert.NoError(err)

	// groth16 zkSNARK: Setup
	innerPK, innerVK, err := groth16.Setup(innerCcs)
	assert.NoError(err)

	// inner witness definition
	innerAssignment := InnerCircuit{
		P: 3,
		Q: 5,
		N: 15,
	}
	innerWitness, err := frontend.NewWitness(&innerAssignment, field)
	assert.NoError(err)
	innerPubWitness, err := innerWitness.Public()
	assert.NoError(err)

	// inner groth16: Prove & Verify
	innerProof, err := groth16.Prove(innerCcs, innerPK, innerWitness)
	assert.NoError(err)
	err = groth16.Verify(innerProof, innerVK, innerPubWitness)
	assert.NoError(err)
	return innerCcs, innerVK, innerPubWitness, innerProof
}
