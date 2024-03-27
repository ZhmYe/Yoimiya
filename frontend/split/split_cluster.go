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
	然后合并电路Cluster -> (得到n份电路)
	然后按序遍历各份电路，进行Setup -> Prove

***/

// initConstraintSystem 初始化cs
func initConstraintSystem() constraint.ConstraintSystem {
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	return cs_bn254.NewR1CS(opt.Capacity)
}

// todo 这里的extra和forwardOutput需要分开两份，不断更新的用于后面的电路，前面一份用于prove的电路
// buildBottomConstraintSystem 这里由于top电路是暂时的，统一写在这里，只关注bottom
// 这里在bottom不为空的情况下才会调用这个函数
func buildBottomConstraintSystem(
	top []int, bottom []int,
	record *DataRecord, assignment frontend.Circuit,
	forwardOutput []constraint.ExtraValue, extra *[]constraint.ExtraValue) (constraint.ConstraintSystem, error) {
	// 这里得到了此次split的top电路
	// forwardOutput对应top需要传给bottom的wireID
	fmt.Print("	Top Circuit ")
	topCs, err := buildConstraintSystemFromIds(top, record, assignment, forwardOutput, *extra, true)
	if err != nil {
		panic(err)
	}
	newExtra, usedExtra := GetExtra(topCs) // 这里得到top电路的extra，这里事实上并未对extra进行更新,value是unset状态
	for _, e := range newExtra {
		e.SignToSet() // 这里强行设置为isSet
		*extra = append(*extra, e)
	}
	for i, e := range *extra {
		count, isUsed := usedExtra[e.GetWireID()]
		if isUsed {
			//e.Consume(count)
			(*extra)[i].Consume(count)
		}
	}
	fmt.Print("	Bottom Circuit ")
	return buildConstraintSystemFromIds(bottom, record, assignment, forwardOutput, *extra, false)
}
func check(cs constraint.ConstraintSystem, originWireNumber int, cut int, index int) bool {
	internalVaribleNumber := cs.GetNbInternalVariables()
	offset := int(float64(originWireNumber)/float64(cut)) * (index + 1)
	return offset+internalVaribleNumber < originWireNumber
}
func getClusterProof(toProveCs constraint.ConstraintSystem, assignment frontend.Circuit,
	forwardOutput []constraint.ExtraValue, extras *[]constraint.ExtraValue,
	isBottom bool) PackedProof {
	fmt.Println(len(forwardOutput), len(*extras))
	switch _toProveR1cs := toProveCs.(type) {
	case *cs_bn254.R1CS:
		if !isBottom {
			err := frontend.SetNbLeaf(assignment, _toProveR1cs, *extras)
			if err != nil {
				panic(err)
			}
		}
		_toProveR1cs.SetForwardOutput(forwardOutput)
	default:
		panic("Only Support bn254 r1cs now...")
	}
	proof := GetSplitProof(toProveCs, assignment, extras, true)
	return proof
}

// SplitAndCluster todo 这里等待完成
// Cluster的过程中，如果是先把所有的instruction记录下来再合并，这样需要依次记录所有的record
// 考虑是否可以每次都初始化一个cs，然后每次切分将对应的instruction add进去 todo
func SplitAndCluster(cs constraint.ConstraintSystem, assignment frontend.Circuit, cut int) ([]PackedProof, error) {
	proofs := make([]PackedProof, 0)
	toSplitCs := cs
	toProveCs := initConstraintSystem()
	flag, round, index, startTime := true, 0, 0, time.Now()
	extras := make([]constraint.ExtraValue, 0)
	forwardOutput := make([]constraint.ExtraValue, 0)
	// 获取原始电路所有的wire数
	originWireNumber := cs.GetNbPublicVariables() + cs.GetNbSecretVariables() + cs.GetNbInternalVariables()
	//item := make([]int, 0)
	fmt.Println("=================Start Recursive Split=================")
	for {
		if !flag {
			//fmt.Println(toSplitCs.GetNbPublicVariables(), toSplitCs.GetNbSecretVariables(), toSplitCs.GetNbInternalVariables())
			proof := getClusterProof(toSplitCs, assignment, forwardOutput, &extras, true)
			proofs = append(proofs, proof)
			fmt.Println()
			fmt.Println("Total Time: ", time.Since(startTime))
			fmt.Println("=================Finish Recursive Split=================")
			break
		}
		round++
		switch _r1cs := toSplitCs.(type) {
		case *cs_bn254.R1CS:
			structureRoundLog(_r1cs, round)

			top, bottom := _r1cs.Sit.CheckAndGetSubCircuitStageIDs()
			record := NewDataRecord(_r1cs)
			// 这里不需要对top的电路进行prove，通过下半电路的internal变量数进行判断
			_r1cs.UpdateForwardOutput()               // 这里从原电路中获得middle对应的wireIDs
			forwardOutput = _r1cs.GetForwardOutputs() // top需要传给bottom的wireID,在构造bottom时需要用到
			// 如果没有bottom了，说明可以将剩下的合并为最后一个split
			if len(bottom) == 0 {
				flag = false
			} else {
				//fmt.Println("bottom=", len(bottom))
				toSplitCs, err := buildBottomConstraintSystem(
					_r1cs.Sit.GetInstructionIdsFromStageIDs(top), _r1cs.Sit.GetInstructionIdsFromStageIDs(bottom),
					record, assignment, forwardOutput, &extras)
				// 判断需要可以开始证明
				if check(toSplitCs, originWireNumber, cut, index) {
					index++
					proofs = append(proofs, getClusterProof(toProveCs, assignment, forwardOutput, &extras, false))
					toProveCs = initConstraintSystem()
					if index == cut-1 {
						flag = false
					}
				} else {
					switch _toProveR1cs := toProveCs.(type) {
					case *cs_bn254.R1CS:
						for _, iID := range _r1cs.Sit.GetInstructionIdsFromStageIDs(top) {
							pi := record.GetPackedInstruction(iID)
							bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
							_toProveR1cs.AddInstructionInSpilt(bID, unpack(pi, record))
						}
					default:
						panic("Only Support bn254 r1cs now...")
					}

				}
				if err != nil {
					panic(err)
				}
			}
		default:
			panic("Only Support bn254 r1cs now...")
		}
	}
	fmt.Println("Clustering Finish...")
	fmt.Println("Start Generate Proofs")
	return proofs, nil
}
