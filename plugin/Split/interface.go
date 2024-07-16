package Split

import (
	"Yoimiya/Circuit"
	"Yoimiya/backend/groth16"
)

type Groth16Prover interface {
	//Initialize(circuit Circuit.TestCircuit) error
	//compile() (constraint.ConstraintSystem, error)

	Process(circuit Circuit.TestCircuit) ([]groth16.Proof, error)
	Record()
}
