package Split

import (
	"Yoimiya/Circuit"
	"Yoimiya/backend/groth16"
	"Yoimiya/plugin"
)

type Groth16SplitRunner struct {
	record   plugin.PluginRecord
	prover   Prover
	verifier Verifier
	split    int
}

func NewGroth16SplitRunner(s int) Groth16SplitRunner {
	return Groth16SplitRunner{
		record:   plugin.NewPluginRecord(),
		prover:   Prover{},
		verifier: Verifier{},
		split:    s,
	}
}
func (r *Groth16SplitRunner) Prepare(circuit Circuit.TestCircuit) {

}
func (r *Groth16SplitRunner) Process(circuit Circuit.TestCircuit) ([]groth16.Proof, error) {
	proofs := make([]groth16.Proof, 0)
	return proofs, nil
}
