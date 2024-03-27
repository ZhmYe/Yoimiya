package split

import (
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/frontend"
	"fmt"
	"time"
)

/***

	Hints: ZhmYe
	一般的prove流程为Compile -> Setup -> Prove(Solve, Run)
	为了减少内存使用，我们修改为Compile -> Split(得到多份电路)
	然后按序遍历各份电路，进行Setup -> Prove

***/

// SplitAndProve 将传入的电路(constraintSystem)切分为多份，返回所有切出的子电路的proof
func SplitAndProve(cs constraint.ConstraintSystem, assignment frontend.Circuit) ([]PackedProof, error) {
	proofs := make([]PackedProof, 0)
	toSplitCs := cs
	flag := true
	extras := make([]constraint.ExtraValue, 0)
	startTime := time.Now()
	round := 0
	fmt.Println("=================Start Recursive Split=================")
	for {
		if !flag {
			fmt.Println()
			fmt.Println("Total Time: ", time.Since(startTime))
			fmt.Println("=================Finish Recursive Split=================")
			break
		}
		round++
		switch _r1cs := toSplitCs.(type) {
		case *cs_bn254.R1CS:

			//switch _r1cs.Sit.Examine() {
			//case graph.PASS:
			//	log := logger.Logger()
			//	log.Debug().Str("SIT LAYER EXAMINE", "PASS").Msg("YZM DEBUG")
			//	//fmt.Println("Examine PASS...")
			//case graph.HAS_LINK:
			//	panic("Sit Layer Error: HAS_LINK...")
			//case graph.LAYER_UNSET:
			//	panic("Sit Layer Error: LAYER_UNSET...")
			//case graph.SPLIT_ERROR:
			//	panic("Sit Layer Error: SPLIT_ERROR...")
			//}
			structureRoundLog(_r1cs, round)
			//sits, err := trySplit(_r1cs)
			top, bottom := _r1cs.Sit.CheckAndGetSubCircuitStageIDs()
			_r1cs.UpdateForwardOutput() // 这里从原电路中获得middle对应的wireIDs
			forwardOutput := _r1cs.GetForwardOutputs()
			//if err != nil {
			//	panic(err)
			//}
			record := NewDataRecord(_r1cs)
			fmt.Print("	Top Circuit ")
			subCs, err := buildConstraintSystemFromIds(_r1cs.Sit.GetInstructionIdsFromStageIDs(top), record, assignment, forwardOutput, extras, true)
			if err != nil {
				panic(err)
			}

			// 这里加入prove的逻辑，这样top可以丢弃
			// 同时包含加入extra的逻辑
			proof := GetSplitProof(subCs, assignment, &extras, false)
			proofs = append(proofs, proof)
			if len(bottom) == 0 {
				flag = false
			} else {
				//fmt.Println("bottom=", len(bottom))
				fmt.Print("	Bottom Circuit ")
				toSplitCs, err = buildConstraintSystemFromIds(_r1cs.Sit.GetInstructionIdsFromStageIDs(bottom), record, assignment, forwardOutput, extras, false)
				if err != nil {
					panic(err)
				}
			}
		default:
			panic("Only Support bn254 r1cs now...")
		}
	}
	return proofs, nil
}
