package evaluate

import (
	"S-gnark/Config"
	"S-gnark/Record"
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	Circuit4VerifyCircuit2 "S-gnark/evaluate/Circuit4VerifyCircuit"
	"github.com/consensys/gnark-crypto/ecc"
)

const LENGTH = Circuit4VerifyCircuit2.LENGTH

func TestRunTimeInDifferentSplitMethod() {
	logger := NewLogWriter("TestTime/")
	methods := []Config.SPLIT_METHOD{Config.SPLIT_LEVELS, Config.SPLIT_STAGES}
	for _, method := range methods {
		Config.Config.Split = method
		for i := 0; i < 100; i++ {
			var innerCcsArray [LENGTH]constraint.ConstraintSystem
			var innerVKArray [LENGTH]groth16.VerifyingKey
			var innerWitnessArray [LENGTH]witness.Witness
			var innerProofArray [LENGTH]groth16.Proof

			for i := 0; i < LENGTH; i++ {
				innerCcs, innerVK, innerWitness, innerProof := Circuit4VerifyCircuit2.GetInner(ecc.BN254.ScalarField())
				innerCcsArray[i] = innerCcs
				innerVKArray[i] = innerVK
				innerWitnessArray[i] = innerWitness
				innerProofArray[i] = innerProof
			}

			// outer proof
			//outerCcs, outerPK, outerVK, full, public := getCircuitVkWitnessPublic(assert, innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray)

			middleCcs, middlePK, middleVK, middleFull, middlePublic := Circuit4VerifyCircuit2.GetCircuitVkWitnessPublic(innerCcsArray, innerVKArray, innerWitnessArray, innerProofArray)

			middleProof, _ := groth16.Prove(middleCcs, middlePK.(groth16.ProvingKey), middleFull)

			err := groth16.Verify(middleProof, middleVK, middlePublic)
			if err != nil {
				return
			}
		}
		switch method {
		case Config.SPLIT_LEVELS:
			logger.Write("Method=SPLIT_LEVELS")
		case Config.SPLIT_STAGES:
			logger.Write("Method=SPLIT_STAGES")
		}
		logger.Wrap()
		splitTime, RunTime := Record.GlobalRecord.GetTime()
		Record.GlobalRecord.Clear()
		splitTime /= 100
		RunTime /= 100
		logger.Write("\tSplit Time=" + splitTime.String())
		logger.Write("\tRun Time=" + (RunTime - splitTime).String())
		logger.Wrap()
	}
	logger.Finish()
}
