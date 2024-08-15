package Component

import (
	"Yoimiya/backend/groth16"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
	"fmt"
	"sync"
	"time"
)

type SolveInput struct {
	tID     int
	phase   int
	witness witness.Witness
}
type SolveOutput struct {
	CommitmentInfo constraint.Groth16Commitments
	Solution       *cs.R1CSSolution
	NbPublic       int
	NbPrivate      int
	tID            int
	phase          int
	PublicWitness  witness.Witness
}
type SolveEngine struct {
	ccs        *cs.R1CS
	input      *chan SolveInput
	output     *chan SolveOutput
	solveLimit int
}

func (se *SolveEngine) Solve(pk groth16.ProvingKey) {
	for input := range *se.input {
		var wg sync.WaitGroup
		wg.Add(se.solveLimit)
		for i := 0; i < se.solveLimit; i++ {
			tmp := input
			go func(input SolveInput) {
				//prover := plugin.NewProver(se.pk)
				startTime := time.Now()
				commitmentsInfo, solution, nbPublic, nbPrivate, err := groth16_bn254.Solve(se.ccs, input.witness, pk.(*groth16_bn254.ProvingKey))
				if err != nil {
					panic(err)
				}
				fmt.Printf("%d solveTime: %s\n", input.tID, time.Since(startTime))
				publicW, err := input.witness.Public()
				if err != nil {
					panic(err)
				}
				*se.output <- SolveOutput{
					CommitmentInfo: commitmentsInfo,
					Solution:       solution,
					NbPublic:       nbPublic,
					NbPrivate:      nbPrivate,
					tID:            input.tID,
					phase:          input.phase,
					PublicWitness:  publicW,
				}
				wg.Done()
			}(tmp)
		}
		wg.Wait()
	}
	close(*se.output)
}
