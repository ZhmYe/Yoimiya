package split

import (
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/frontend"
	"S-gnark/graph/Sit"
	"fmt"
	"time"
)

/*
**

	Hints: ZhmYe
	一般的prove流程为Compile -> Setup -> Prove(Solve, Run)
	为了减少内存使用，我们修改为Compile -> Split(得到多份电路)
	然后合并电路Cluster -> (得到n份电路)
	然后按序遍历各份电路，进行Setup -> Prove

**
*/

// ProveConstraintSystem 待证明的电路
type ProveConstraintSystem struct {
	cs            constraint.ConstraintSystem
	extra         []constraint.ExtraValue // 这是从前面的proveCs里继承的
	forwardOutput []constraint.ExtraValue // 这是要传递给后面的
	assignment    frontend.Circuit
	// 因此， 下一个toProveCs的extra应该是extra + forwardOutput
}

func NewProveConstraintSystem(assignment frontend.Circuit) *ProveConstraintSystem {
	cs := initConstraintSystem()
	return &ProveConstraintSystem{
		cs:            cs,
		extra:         make([]constraint.ExtraValue, 0),
		forwardOutput: make([]constraint.ExtraValue, 0),
		assignment:    assignment,
	}
}
func (p *ProveConstraintSystem) SetAssignment(assignment frontend.Circuit) {
	p.assignment = assignment
}
func (p *ProveConstraintSystem) SetExtra(extra []constraint.ExtraValue) {
	p.extra = extra
	toProveCs := p.cs
	switch _toProveR1cs := toProveCs.(type) {
	case *cs_bn254.R1CS:
		// 这里不更新extra
		err := frontend.SetNbLeaf(p.assignment, _toProveR1cs, p.extra)
		if err != nil {
			panic(err)
		}
	default:
		panic("Only Support bn254 r1cs now...")
	}
}
func (p *ProveConstraintSystem) AddForwardOutput(forwardOutput []constraint.ExtraValue) {
	p.forwardOutput = append(p.forwardOutput, forwardOutput...)
}
func (p *ProveConstraintSystem) Forward() []constraint.ExtraValue {
	//fmt.Println(len(p.extra), len(p.forwardOutput))
	return append(p.extra, p.forwardOutput...)
}

//// SplitConstraintSystem 待切分的电路
//type SplitConstraintSystem struct {
//	cs    constraint.ConstraintSystem
//	extra []constraint.ExtraValue
//}
//func NewSplitConstraintSystem() *SplitConstraintSystem {
//
//}

// initConstraintSystem 初始化cs
func initConstraintSystem() constraint.ConstraintSystem {
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	//cs := cs_bn254.NewR1CS(opt.Capacity)
	//cs.CoeffTable = record.GetCoeffTable()
	return cs_bn254.NewR1CS(opt.Capacity)
}

// buildBottomConstraintSystem 这里由于top电路是暂时的，统一写在这里，只关注bottom
// 这里在bottom不为空的情况下才会调用这个函数
func buildBottomConstraintSystem(
	bottom []int,
	record *DataRecord, assignment frontend.Circuit,
	toProveCs ProveConstraintSystem) (constraint.ConstraintSystem, error) {
	// 这里根据toProve的forwardOutput + extra来构建bottom电路
	opt := frontend.DefaultCompileConfig()
	//fmt.Println("capacity=", opt.Capacity)
	cs := cs_bn254.NewR1CS(opt.Capacity)
	//if isTop {
	//	SetForwardOutput(cs, forwardOutput) // 设置应该传到bottom的wireID
	//}
	// 下半电路的extra就是toProve的forwardOutput + extra
	err := frontend.SetNbLeaf(assignment, cs, toProveCs.Forward())
	if err != nil {
		return nil, err
	}
	//fmt.Println("nbPublic=", cs.GetNbPublicVariables(), " nbPrivate=", cs.GetNbSecretVariables())
	for _, iID := range bottom {
		pi := record.GetPackedInstruction(iID)
		bID := cs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
		cs.AddInstructionInSpilt(bID, unpack(pi, record))
		//// 由于instruction变化，所以在这里需要重新映射stage内部的iID
		//sit.ModifyiID(i, j, len(cs.Instructions)) // 这里是串行添加的，新的Instruction id就是当前的长度
	}
	cs.CoeffTable = record.GetCoeffTable()
	fmt.Println("	ToSplitCs Result: ")
	fmt.Println("		NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("		NbCoeff=", cs.GetNbConstraints())
	fmt.Println("		NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
	return cs, nil
}
func check(cs constraint.ConstraintSystem, originWireNumber int, cut int, index int) bool {
	VariableNumber := cs.GetNbPublicVariables() + cs.GetNbSecretVariables() + cs.GetNbInternalVariables()
	offset := int(float64(originWireNumber) / float64(cut))
	return VariableNumber >= offset
}

func getClusterProof(toProve *ProveConstraintSystem, assignment frontend.Circuit) PackedProof {
	toProveCs := toProve.cs
	switch _toProveR1cs := toProveCs.(type) {
	case *cs_bn254.R1CS:
		// 这里不更新extra
		//err := frontend.SetNbLeaf(assignment, _toProveR1cs, toProve.extra)
		//if err != nil {
		//	panic(err)
		//}
		_toProveR1cs.SetForwardOutput(toProve.forwardOutput)
	default:
		panic("Only Support bn254 r1cs now...")
	}
	// 这里会更新extra
	proof := GetSplitProof(toProveCs, assignment, &(toProve.extra), true)
	return proof
}

func getGlobalRecord(cs constraint.ConstraintSystem) *DataRecord {
	switch _r1cs := cs.(type) {
	case *cs_bn254.R1CS:
		return NewDataRecord(_r1cs)
	default:
		panic("Only Support bn254 r1cs now...")
	}
}

func updateProveCs(toProve *ProveConstraintSystem, iIDs []int, record *DataRecord) {
	toProveCs := toProve.cs
	switch _toProveR1cs := toProveCs.(type) {
	case *cs_bn254.R1CS:
		_toProveR1cs.CoeffTable = record.CoeffTable
		for _, iID := range iIDs {
			pi := record.GetPackedInstruction(iID)
			bID := toProveCs.AddBlueprint(record.GetBluePrint(pi.BlueprintID))
			_toProveR1cs.AddInstructionInSpilt(bID, unpack(pi, record))
		}
		fmt.Println("	ToProveCs Update: ")
		fmt.Println("		NbPublic=", toProveCs.GetNbPublicVariables(), " NbSecret=", toProveCs.GetNbSecretVariables(), " NbInternal=", toProveCs.GetNbInternalVariables())
		fmt.Println("		NbCoeff=", toProveCs.GetNbConstraints())
		fmt.Println("		NbWires=", toProveCs.GetNbPublicVariables()+toProveCs.GetNbSecretVariables()+toProveCs.GetNbInternalVariables())

	default:
		panic("Only Support bn254 r1cs now...")
	}
}

// SplitAndCluster todo 这里等待完成
// Cluster的过程中，如果是先把所有的instruction记录下来再合并，这样需要依次记录所有的record
// 考虑是否可以每次都初始化一个cs，然后每次切分将对应的instruction add进去 todo
func SplitAndCluster(cs constraint.ConstraintSystem, assignment frontend.Circuit, cut int) ([]PackedProof, error) {

	//globalRecord := getGlobalRecord(cs) // 这里记录全局的record，所有的电路都从这里走
	var record *DataRecord
	proofs := make([]PackedProof, 0)
	toSplitCs := cs
	var sit *Sit.SITree
	flag, round, index, startTime := true, 0, 0, time.Now()
	extras := make([]constraint.ExtraValue, 0)
	forwardOutput := make([]constraint.ExtraValue, 0)
	toProveCs := NewProveConstraintSystem(assignment)
	// 第一个proveCs的extras是空的
	toProveCs.SetExtra(extras) // forwardOutput在每次切分电路的top时将top的forwardOutput添加
	// 获取原始电路所有的wire数
	originWireNumber := cs.GetNbPublicVariables() + cs.GetNbSecretVariables() + cs.GetNbInternalVariables()
	//item := make([]int, 0)
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
			structureRoundLog(_r1cs, round)
			_r1cs.UpdateForwardOutput()
			record = NewDataRecord(_r1cs)
			// 这里从待切电路中获得middle对应的wireIDs，用于传给下一个toSplitCs
			forwardOutput = _r1cs.GetForwardOutputs() // 得到本次待切电路的上半电路需要传给下半电路的forwardOutput
			toProveCs.AddForwardOutput(forwardOutput) // 添加proveCs的forwardOutput
			switch _sit := _r1cs.SplitEngine.(type) {
			case *Sit.SITree:
				sit = _sit
			default:
				panic("Only Support SIT now...")
			}
			//sit = _r1cs.SplitEngine.(type)
		default:
			panic("Only Support bn254 r1cs now...")
		}
		top, bottom := sit.CheckAndGetSubCircuitStageIDs() // 得到toSplitCs再次切分后的top和bottom
		if len(bottom) == 0 {
			// 如果已经无法再切分，那么toSplitCs就是最后的电路
			// todo 这里的逻辑还需要更新
			// 这里将toSplitCs的所有instruction更新到toProveCs中
			flag = false
			switch finalCs := toSplitCs.(type) {
			case *cs_bn254.R1CS:
				updateProveCs(toProveCs, finalCs.SplitEngine.GetAllInstructions(), record)
				proof := getClusterProof(toProveCs, assignment)
				proofs = append(proofs, proof)
			default:
				panic("Only Support bn254 r1cs now...")
			}
		} else {
			// 如果还可以继续切分，那么将top部分加入到toProveCs中，然后用bottom构建新的toSplitCs
			// 对toProveCs的cs进行更新
			updateProveCs(toProveCs, sit.GetInstructionIdsFromStageIDs(top), record)
			var err error
			toSplitCs, err = buildBottomConstraintSystem(
				sit.GetInstructionIdsFromStageIDs(bottom),
				record, assignment, *toProveCs)
			if err != nil {
				panic(err)
			}
			// 如果上半电路数量足够,internal数量超过总数除以份数
			if check(toProveCs.cs, originWireNumber, cut, index) {
				// 那么toProveCs可以进行证明，
				//toProveCs的extra应该在toProveCs被创建时就被确定即从前面的toProveCs中继承了哪些wire
				proof := getClusterProof(toProveCs, assignment)
				proofs = append(proofs, proof)
				// 创建新的toProve，此时得到上一个toProve的forwardOutput和extra并进行合并
				// 这部分逻辑写在toProve里
				nextExtra := toProveCs.Forward()
				toProveCs = NewProveConstraintSystem(assignment)
				toProveCs.SetExtra(nextExtra)
			} else {
				// 如果还不能进行证明，没有任何额外逻辑，继续进行后续的split逻辑
			}
			if len(proofs) == cut {
				flag = false
			}
		}
	}
	fmt.Println("Clustering Finish...")
	fmt.Println("Start Generate Proofs")
	return proofs, nil
}
