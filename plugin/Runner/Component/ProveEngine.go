package Component

import (
	"Yoimiya/backend/groth16"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/frontend/split"
	"runtime"
	"sync"
)

type ProveOutput struct {
	proof split.PackedProof
	tID   int
	phase int
}
type ProveEngine struct {
	input      *chan SolveOutput
	output     *chan ProveOutput
	proveLimit int
}

func (pe *ProveEngine) Prove(pk groth16.ProvingKey, vk groth16.VerifyingKey) {
	for input := range *pe.input {
		var wg sync.WaitGroup
		wg.Add(pe.proveLimit)
		for i := 0; i < pe.proveLimit; i++ {
			tmp := input
			go func(input SolveOutput) {
				//startTime := time.Now()
				proof, err := groth16_bn254.GenerateZKP(input.CommitmentInfo, *input.Solution, pk.(*groth16_bn254.ProvingKey), input.NbPublic, input.NbPrivate)
				if err != nil {
					panic(err)
				}
				//fmt.Printf("%d ProveTime: %s\n", input.tID, time.Since(startTime))
				//publicWitness, err := witness.Public()
				if err != nil {
					panic(err)
				}
				*pe.output <- ProveOutput{
					proof: split.NewPackedProof(proof, vk, input.PublicWitness),
					tID:   input.tID,
					phase: input.phase,
				}
				runtime.GC()
				wg.Done()
				//t.proofs = append(t.proofs, split.NewPackedProof(proof, vk, input.PublicWitness))
			}(tmp)
		}
		wg.Wait()
	}
	close(*pe.output)
}
