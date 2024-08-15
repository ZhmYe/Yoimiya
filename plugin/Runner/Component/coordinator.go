package Component

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin/Runner"
	"github.com/consensys/gnark-crypto/ecc"
)

type Coordinator struct {
	solver     SolveEngine
	prover     ProveEngine
	solverPool chan SolveInput
	proverPool chan SolveOutput
	output     chan ProveOutput
	pk         groth16.ProvingKey
	vk         groth16.VerifyingKey
	inputID    []int
}

func NewCoordinator(ccs constraint.ConstraintSystem, pk groth16.ProvingKey, vk groth16.VerifyingKey, inputID []int, solveLimit int, proveLimit int) *Coordinator {
	solverPool := make(chan SolveInput, 100)
	proverPool := make(chan SolveOutput, 100)
	output := make(chan ProveOutput)
	return &Coordinator{
		solver: SolveEngine{
			ccs:        ccs.(*cs.R1CS),
			input:      &solverPool,
			output:     &proverPool,
			solveLimit: solveLimit,
		},
		prover: ProveEngine{
			input:      &proverPool,
			output:     &output,
			proveLimit: proveLimit,
		},
		solverPool: solverPool,
		proverPool: proverPool,
		output:     output,
		pk:         pk,
		vk:         vk,
		inputID:    inputID,
	}
}
func (c *Coordinator) Inject(tasks []*Runner.Task) {
	for _, task := range tasks {
		request := task.Params()
		witness, err := frontend.GenerateSplitWitnessFromPli(request.Pli, c.inputID, request.Extra, ecc.BN254.ScalarField())
		if err != nil {
			panic(err)
		}
		c.solverPool <- SolveInput{
			tID:     request.ID,
			phase:   request.Phase,
			witness: witness,
		}
	}
	close(c.solverPool)
}
func (c *Coordinator) Process(tasks []*Runner.Task) {
	go c.solver.Solve(c.pk)
	go c.prover.Prove(c.pk, c.vk)
	for o := range c.output {
		tID := o.tID
		tasks[tID].HandleResponse(Runner.Response{
			Proof: o.proof,
			Extra: make([]constraint.ExtraValue, 0),
			ID:    tID,
			Phase: o.phase,
		})
	}

}
