package constraint

import (
	"Yoimiya"
	"Yoimiya/Config"
	"Yoimiya/constraint/solver"
	"Yoimiya/debug"
	"Yoimiya/internal/tinyfield"
	"Yoimiya/internal/utils"
	"Yoimiya/logger"
	"Yoimiya/profile"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"math/big"
	"strconv"
	"sync"
)

type SystemType uint16

const (
	SystemUnknown SystemType = iota
	SystemR1CS
	SystemSparseR1CS
)

// PackedInstruction is the lowest element of a constraint system. It stores just enough data to
// reconstruct a constraint of any shape or a hint at solving time.
type PackedInstruction struct {
	// BlueprintID maps this instruction to a blueprint
	BlueprintID BlueprintID

	// ConstraintOffset stores the starting constraint ID of this instruction.
	// Might not be strictly necessary; but speeds up solver for instructions that represents
	// multiple constraints.
	ConstraintOffset uint32

	// WireOffset stores the starting internal wire ID of this instruction. Blueprints may use this
	// and refer to output wires by their offset.
	// For example, if a blueprint declared 5 outputs, the first output wire will be WireOffset,
	// the last one WireOffset+4.
	WireOffset uint32

	// The constraint system stores a single []uint32 calldata slice. StartCallData
	// points to the starting index in the mentioned slice. This avoid storing a slice per
	// instruction (3 * uint64 in memory).
	StartCallData uint64
}

// Unpack returns the instruction corresponding to the packed instruction.
func (pi PackedInstruction) Unpack(cs *System) Instruction {

	blueprint := cs.Blueprints[pi.BlueprintID]
	cSize := blueprint.CalldataSize()
	if cSize < 0 {
		// by convention, we store nbInputs < 0 for non-static input length.
		cSize = int(cs.CallData[pi.StartCallData])
	}

	return Instruction{
		ConstraintOffset: pi.ConstraintOffset,
		WireOffset:       pi.WireOffset,
		Calldata:         cs.CallData[pi.StartCallData : pi.StartCallData+uint64(cSize)],
	}
}

// Instruction is the lowest element of a constraint system. It stores all the data needed to
// reconstruct a constraint of any shape or a hint at solving time.
type Instruction struct {
	ConstraintOffset uint32
	WireOffset       uint32
	Calldata         []uint32
}

/*** Hints: ZhmYe
	Here can add new structures
	todo: modify
***/

// add by ZhmYe

// ExtraValue 用来表示由前置电路的输出作为后置电路的输入的所有wire
type ExtraValue struct {
	wireID int
	value  fr.Element
	isSet  bool
	count  int // 用于记录该ExtraValue还剩多少次被用机会
}

func NewExtraValue(wireID int) ExtraValue {
	return ExtraValue{
		wireID: wireID,
		value:  fr.Element{},
		isSet:  false,
		//count:  0,
	}
}
func (v *ExtraValue) SignToSet() {
	v.isSet = true
}
func (v *ExtraValue) SetValue(value fr.Element) {
	v.value = value
	v.isSet = true
}

//func (v *ExtraValue) SetCount(count int) {
//	v.count = count
//}
//func (v *ExtraValue) GetCount() int {
//	return v.count
//}

//func (v *ExtraValue) Consume(count int) {
//	v.count -= count
//	if v.count < 0 {
//		panic("error")
//	}
//}

//func (v *ExtraValue) IsUsed() bool {
//	return v.count == 0
//}

func (v *ExtraValue) GetWireID() int {
	return v.wireID
}
func (v *ExtraValue) GetValue() fr.Element {
	if !v.isSet {
		panic("value not set...")
	}
	return v.value
}

// System contains core elements for a constraint System
type System struct {
	// serialization header
	GnarkVersion string
	ScalarField  string

	Type SystemType

	Instructions []PackedInstruction
	Blueprints   []Blueprint
	CallData     []uint32 // huge slice.

	// can be != than len(instructions)
	NbConstraints int

	// number of internal wires
	NbInternalVariables int

	// input wires names
	Public, Secret []string

	// logs (added with system.Println, resolved when solver sets a value to a wire)
	Logs []LogEntry

	// debug info contains stack trace (including line number) of a call to a system.API that
	// results in an unsolved constraint
	DebugInfo   []LogEntry
	SymbolTable debug.SymbolTable
	// maps constraint id to debugInfo id
	// several constraints may point to the same debug info
	MDebug map[int]int

	// maps hintID to hint string identifier
	MHintsDependencies map[solver.HintID]string

	// each level contains independent constraints and can be parallelized
	// it is guaranteed that all dependencies for constraints in a level l are solved
	// in previous levels
	// TODO @gbotrel these are currently updated after we add a constraint.
	// but in case the object is built from a serialized representation
	// we need to init the level builder lbWireLevel from the existing constraints.
	Levels [][]int

	// scalar field
	q      *big.Int `cbor:"-"`
	bitLen int      `cbor:"-"`

	// level builder
	lbWireLevel []Level `cbor:"-"` // at which level we solve a wire. init at -1.

	CommitmentInfo Commitments
	GkrInfo        GkrInfo

	genericHint BlueprintID

	// add by ZhmYe
	Wires2Instruction map[uint32]int // an output wire "w" first compute in Instruction "I", then store "w" -> "i"
	Bias              map[uint32]int // 记录wireID在切割后的电路里对应的下标
	// 这些是自底向上建SIT所需要的
	//InstructionForwardDAG  *graph.DAG     // DAG constructed by Instructions, forward
	//InstructionBackwardDAG *graph.DAG     // DAG constructed by Instructions, backward
	//degree                 map[int]int    // store each node's degree(to order)
	SplitEngine SplitEngine
	//Sit           *Sit.SITree
	forwardOutput []ExtraValue // 这里用来需要传下去的wireID
	extra         int          // 这里用来记录extra的数量，用于确认原本的public variable有多少个
	//usedExtra     map[int]int  // 记录extra值被使用的次数
	//extra         []fr.Element // 这里用来记录
	deepest []int // 这里用来记录每个instruction的后续Instruction最大抵达的深度
}

// NewSystem initialize the common structure among constraint system
func NewSystem(scalarField *big.Int, capacity int, t SystemType) System {
	system := System{
		Type:               t,
		SymbolTable:        debug.NewSymbolTable(),
		MDebug:             map[int]int{},
		GnarkVersion:       gnark.Version.String(),
		ScalarField:        scalarField.Text(16),
		MHintsDependencies: make(map[solver.HintID]string),
		q:                  new(big.Int).Set(scalarField),
		bitLen:             scalarField.BitLen(),
		Instructions:       make([]PackedInstruction, 0, capacity),
		CallData:           make([]uint32, 0, capacity*8),
		lbWireLevel:        make([]Level, 0, capacity),
		Levels:             make([][]int, 0, capacity/2),
		CommitmentInfo:     NewCommitments(t),

		// add by ZhmYe
		Wires2Instruction: make(map[uint32]int),
		//InstructionForwardDAG:  graph.NewDAG(),
		//InstructionBackwardDAG: graph.NewDAG(),
		//degree:                 make(map[int]int),
		SplitEngine: InitSplitEngine(),
		//Sit:           Sit.NewSITree(),
		forwardOutput: make([]ExtraValue, 0),
		//extra:         make([]fr.Element, 0),
		Bias:  make(map[uint32]int),
		extra: 0,
		//usedExtra: make(map[int]int), // 现在只有两半电路，暂时先不要
	}

	system.genericHint = system.AddBlueprint(&BlueprintGenericHint{})
	return system
}

// add by ZhmYe

// SetExtraNumber 设置extra的public 数量
func (system *System) SetExtraNumber(extra []ExtraValue) {
	system.extra = len(extra)
}

// GetExtraNumber 获取extra的public 数量
func (system *System) GetExtraNumber() int {
	return system.extra
}

// // UpdateUsedExtra 更新extra的使用记录，在判断变量是否为input的时候调用
//
//	func (system *System) UpdateUsedExtra(wireID int) {
//		_, exist := system.usedExtra[wireID]
//		if !exist {
//			system.usedExtra[wireID] = 1
//		} else {
//			system.usedExtra[wireID]++
//		}
//	}
//
//	func (system *System) GetUsedExtra() map[int]int {
//		return system.usedExtra
//	}

func (system *System) SetExtraValue(idx int, value fr.Element) {
	system.forwardOutput[idx].SetValue(value)
}

// SetBias 用于设置Bias, wireID -> idx
func (system *System) SetBias(wireID uint32, idx int) {
	bias, exist := system.Bias[wireID]
	if exist {
		//fmt.Println("wireID = ", wireID, ", bias = ", bias, " , offset = ", system.internalWireOffset())
		if bias < int(system.internalWireOffset()) {
			return
		}
		panic("Wire Bias Has been Set")
	}
	system.Bias[wireID] = idx
	//if idx+1 != system.GetNbSecretVariables()+system.GetNbPublicVariables()+system.GetNbInternalVariables() {
	//	fmt.Println("Set Bias Error!!!")
	//}
}
func (system *System) GetBias() map[uint32]int {
	return system.Bias
}
func (system *System) GetWireBias(wireID int) int {
	bias, exist := system.Bias[uint32(wireID)]
	if exist {
		return bias
	}
	return wireID
}

// SetForwardOutput 这里把上一个电路传下来的WireID记录在system.forwardOutput
func (system *System) SetForwardOutput(output []ExtraValue) {
	//for _, wireID := range output {
	//	system.forwardOutput = append(system.forwardOutput, NewExtraValue(wireID))
	//}
	system.forwardOutput = output
}
func (system *System) GetNbForwardOutput() int {
	return len(system.forwardOutput)
}

//	func (system *System) GetForwardOutput(wireID int) int {
//		return system.forwardOutput[index]
//	}

func (system *System) GetForwardOutputs() []ExtraValue {
	return system.forwardOutput
}
func (system *System) GetForwardOutputIds() []int {
	result := make([]int, 0)
	for _, e := range system.forwardOutput {
		result = append(result, e.GetWireID())
	}
	return result
}

//func (system *System) AddExtra(value fr.Element) {
//	system.extra = append(system.extra, value)
//}
//func (system *System) GetExtra() []fr.Element {
//	return system.extra
//}

// add by ZhmYe

// UpdateForwardOutput 这里返回传给下一个电路的 WireID
func (system *System) UpdateForwardOutput() {
	middleInstructions := system.SplitEngine.GetMiddleOutputs()
	// 这里我们要得到Instruction里的所有WireID
	// todo 是否要所有的WireID? 还是只要output
	// 这里暂时按照只要output来写
	// 首先先要得到Instruction
	//system.forwardOutput = make([]frontend.ExtraValue, 0)
	// system.Wire2Instruction记录了Wire在哪一个Instruction中作为output
	for wire, iID := range system.Wires2Instruction {
		_, isMiddle := middleInstructions[iID]
		if isMiddle {
			e := NewExtraValue(int(wire))
			//e.SetCount(count) // 记录该extra会被使用多少次
			system.forwardOutput = append(system.forwardOutput, e)
		}
	}
	//for _, wireID := range system.GetForwardOutputs() {
	//	system.AddExtra(Asolver.GetWireValue(wireID))
	//}
}

// AppendWire2Instruction add by ZhmYe
func (system *System) AppendWire2Instruction(wireId uint32, iID int) {
	system.Wires2Instruction[wireId] = iID
}

// GetDegree add by ZhmYe
//func (system *System) GetDegree(id int) int {
//	d, exist := system.degree[id]
//	if !exist {
//		return -1
//	} else {
//		return d
//	}
//}
//func (system *System) initDegree(id int) {
//	_, exist := system.degree[id]
//	if !exist {
//		system.degree[id] = 0
//	}
//}
//
//// UpdateDegree add by ZhmYe
//func (system *System) UpdateDegree(sub bool, ids ...int) {
//	for _, id := range ids {
//		_, exist := system.degree[id]
//		if !exist && sub {
//			// 没有这个内容无法减一
//			fmt.Errorf("can't update sub operation to an undefined id")
//			return
//		}
//		if !exist && !sub {
//			system.degree[id] = 1
//		}
//		if exist {
//			if sub {
//				system.degree[id]--
//			} else {
//				system.degree[id]++
//			}
//		}
//	}
//}
//
//// GetZeroDegree add by ZhmYe
//func (system *System) GetZeroDegree() (result []int) {
//	for id, d := range system.degree {
//		if d == 0 {
//			result = append(result, id)
//		}
//	}
//	return result
//}
//
//// GetDAGs add by ZhmYe
//func (system *System) GetDAGs() (*graph.DAG, *graph.DAG) {
//	return system.InstructionForwardDAG, system.InstructionBackwardDAG
//}
//
//// GetOrder add by ZhmYe
//// 相当于对DAG进行拓扑排序
//// 这里由于后续修改，需要把degree修改回前向的degree才可以使用，这里的拓扑排序分层结果等价于源代码中的Levels
//func (system *System) GetOrder() (order [][]int) {
//	for {
//		zeroList := system.GetZeroDegree()
//		if len(zeroList) == 0 {
//			break
//		}
//		order = append(order, zeroList)
//		for _, id := range zeroList {
//			links := system.InstructionForwardDAG.GetLinks(id)
//			links = append(links, id)
//			system.UpdateDegree(true, links...)
//		}
//	}
//	//fmt.Println(len(order))
//	return order
//}

// GetNbInstructions returns the number of instructions in the system
func (system *System) GetNbInstructions() int {
	return len(system.Instructions)
}

// GetInstruction returns the instruction at index id
func (system *System) GetInstruction(id int) Instruction {
	return system.Instructions[id].Unpack(system)
}

// AddBlueprint adds a blueprint to the system and returns its ID
func (system *System) AddBlueprint(b Blueprint) BlueprintID {
	system.Blueprints = append(system.Blueprints, b)
	return BlueprintID(len(system.Blueprints) - 1)
}

func (system *System) GetNbSecretVariables() int {
	return len(system.Secret)
}
func (system *System) GetNbPublicVariables() int {
	return len(system.Public)
}
func (system *System) GetNbInternalVariables() int {
	return system.NbInternalVariables
}
func (system *System) GetNbWires() int {
	return system.GetNbSecretVariables() + system.GetNbPublicVariables() + system.GetNbInternalVariables()
}

// CheckSerializationHeader parses the scalar field and gnark version headers
//
// This is meant to be use at the deserialization step, and will error for illegal values
func (system *System) CheckSerializationHeader() error {
	// check gnark version
	binaryVersion := gnark.Version
	objectVersion, err := semver.Parse(system.GnarkVersion)
	if err != nil {
		return fmt.Errorf("when parsing gnark version: %w", err)
	}

	if binaryVersion.Compare(objectVersion) != 0 {
		log := logger.Logger()
		log.Warn().Str("binary", binaryVersion.String()).Str("object", objectVersion.String()).Msg("gnark version (binary) mismatch with constraint system. there are no guarantees on compatibility")
	}

	// TODO @gbotrel maintain version changes and compare versions properly
	// (ie if major didn't change,we shouldn't have a compatibility issue)

	scalarField := new(big.Int)
	_, ok := scalarField.SetString(system.ScalarField, 16)
	if !ok {
		return fmt.Errorf("when parsing serialized modulus: %s", system.ScalarField)
	}
	curveID := utils.FieldToCurve(scalarField)
	if curveID == ecc.UNKNOWN && scalarField.Cmp(tinyfield.Modulus()) != 0 {
		return fmt.Errorf("unsupported scalar field %s", scalarField.Text(16))
	}
	system.q = new(big.Int).Set(scalarField)
	system.bitLen = system.q.BitLen()
	return nil
}

// GetNbVariables return number of internal, secret and public variables
func (system *System) GetNbVariables() (internal, secret, public int) {
	return system.NbInternalVariables, system.GetNbSecretVariables(), system.GetNbPublicVariables()
}

func (system *System) Field() *big.Int {
	return new(big.Int).Set(system.q)
}

// bitLen returns the number of bits needed to represent a fr.Element
func (system *System) FieldBitLen() int {
	return system.bitLen
}

func (system *System) AddInternalVariable() (idx int) {
	idx = system.NbInternalVariables + system.GetNbPublicVariables() + system.GetNbSecretVariables()
	system.NbInternalVariables++
	//fmt.Println(system.NbInternalVariables)
	// also grow the level slice
	// modify by ZhmYe
	if Config.Config.Split == Config.SPLIT_LEVELS {
		system.lbWireLevel = append(system.lbWireLevel, LevelUnset)
		if debug.Debug && len(system.lbWireLevel) != system.NbInternalVariables {
			panic("internal error")
		}
	}
	return idx
}

func (system *System) AddPublicVariable(name string) (idx int) {
	idx = system.GetNbPublicVariables()
	system.Public = append(system.Public, name)
	return idx
}

func (system *System) AddSecretVariable(name string) (idx int) {
	idx = system.GetNbSecretVariables() + system.GetNbPublicVariables()
	system.Secret = append(system.Secret, name)
	return idx
}

func (system *System) AddSolverHint(f solver.Hint, id solver.HintID, input []LinearExpression, nbOutput int) (internalVariables []int, err error) {
	if nbOutput <= 0 {
		return nil, fmt.Errorf("hint function must return at least one output")
	}

	var name string
	if id == solver.GetHintID(f) {
		name = solver.GetHintName(f)
	} else {
		name = strconv.Itoa(int(id))
	}

	// register the hint as dependency
	if registeredName, ok := system.MHintsDependencies[id]; ok {
		// hint already registered, let's ensure string registeredName matches
		if registeredName != name {
			return nil, fmt.Errorf("hint dependency registration failed; %s previously register with same UUID as %s", name, registeredName)
		}
	} else {
		system.MHintsDependencies[id] = name
	}

	// prepare wires
	internalVariables = make([]int, nbOutput)
	for i := 0; i < len(internalVariables); i++ {
		internalVariables[i] = system.AddInternalVariable()
		//// todo 这里似乎wireID就是下标？
		//system.SetBias(uint32(internalVariables[i]), internalVariables[i])
	}

	// associate these wires with the solver hint
	hm := HintMapping{
		HintID: id,
		Inputs: input,
		OutputRange: struct {
			Start uint32
			End   uint32
		}{
			uint32(internalVariables[0]),
			uint32(internalVariables[len(internalVariables)-1]) + 1,
		},
	}

	blueprint := system.Blueprints[system.genericHint]

	// get []uint32 from the pool
	calldata := getBuffer()

	blueprint.(BlueprintHint).CompressHint(hm, calldata)

	system.AddInstruction(system.genericHint, *calldata)

	// return []uint32 to the pool
	putBuffer(calldata)

	return
}

func (system *System) AddCommitment(c Commitment) error {
	switch v := c.(type) {
	case Groth16Commitment:
		system.CommitmentInfo = append(system.CommitmentInfo.(Groth16Commitments), v)
	case PlonkCommitment:
		system.CommitmentInfo = append(system.CommitmentInfo.(PlonkCommitments), v)
	default:
		return fmt.Errorf("unknown commitment type %T", v)
	}
	return nil
}

// GetCommitmentInfoInSplit add by ZhmYe
// 用于获取bias后的commitment
func (system *System) GetCommitmentInfoInSplit() Groth16Commitments {
	commitmentInfo := system.CommitmentInfo.(Groth16Commitments)
	NewCommitmentInfo := make(Groth16Commitments, 0)
	for _, commitment := range commitmentInfo {
		// 首先修改commitmentIndex
		commitmentIndex := system.GetWireBias(commitment.CommitmentIndex)
		publicAndCommitmentCommitted := make([]int, len(commitment.PublicAndCommitmentCommitted))
		for j, id := range commitment.PublicAndCommitmentCommitted {
			publicAndCommitmentCommitted[j] = system.GetWireBias(id)
		}
		privateCommitted := make([]int, len(commitment.PrivateCommitted))
		for k, id := range commitment.PrivateCommitted {
			privateCommitted[k] = system.GetWireBias(id)
		}
		NewCommitmentInfo = append(NewCommitmentInfo, Groth16Commitment{
			PublicAndCommitmentCommitted: publicAndCommitmentCommitted,
			PrivateCommitted:             privateCommitted,
			CommitmentIndex:              commitmentIndex,
			NbPublicCommitted:            commitment.NbPublicCommitted,
		})
	}
	return NewCommitmentInfo

}

func (system *System) AddLog(l LogEntry) {
	system.Logs = append(system.Logs, l)
}

func (system *System) AttachDebugInfo(debugInfo DebugInfo, constraintID []int) {
	system.DebugInfo = append(system.DebugInfo, LogEntry(debugInfo))
	id := len(system.DebugInfo) - 1
	for _, cID := range constraintID {
		system.MDebug[cID] = id
	}
}

// VariableToString implements Resolver
func (system *System) VariableToString(vID int) string {
	nbPublic := system.GetNbPublicVariables()
	nbSecret := system.GetNbSecretVariables()

	if vID < nbPublic {
		return system.Public[vID]
	}
	vID -= nbPublic
	if vID < nbSecret {
		return system.Secret[vID]
	}
	vID -= nbSecret
	return fmt.Sprintf("v%d", vID) // TODO @gbotrel  vs strconv.Itoa.
}

/***
	Hints: ZhmYe
	Here call "AddInstruction"
***/

func (cs *System) AddR1C(c R1C, bID BlueprintID) int {
	profile.RecordConstraint()
	blueprint := cs.Blueprints[bID]

	// get a []uint32 from a pool
	calldata := getBuffer()

	// compress the R1C into a []uint32 and add the instruction
	blueprint.(BlueprintR1C).CompressR1C(&c, calldata)
	/***
		Hints: ZhmYe
		callData:
			len: 4 + 2 * len(L) + 2 * len(R) + 2 * len(O)
			4: len, len(L), len(R), len(O)
			2* len(L)/len(R)/len(O): 2 * (CoeffID(), WireID())
	***/
	cs.AddInstruction(bID, *calldata)

	// release the []uint32 to the pool
	putBuffer(calldata)

	return cs.NbConstraints - 1
}

func (cs *System) AddSparseR1C(c SparseR1C, bID BlueprintID) int {
	profile.RecordConstraint()

	blueprint := cs.Blueprints[bID]

	// get a []uint32 from a pool
	calldata := getBuffer()

	// compress the SparceR1C into a []uint32 and add the instruction
	blueprint.(BlueprintSparseR1C).CompressSparseR1C(&c, calldata)

	cs.AddInstruction(bID, *calldata)

	// release the []uint32 to the pool
	putBuffer(calldata)

	return cs.NbConstraints - 1
}

/***

	Hints: ZhmYe
	todo
	1. 现在的代码是怎么划分level的，目前level内部独立，level之间可能相连
	2. 如何修改？

***/

func (cs *System) AddInstructionInSpilt(bID BlueprintID, calldata []uint32) []uint32 {
	// set the offsets
	pi := PackedInstruction{
		StartCallData:    uint64(len(cs.CallData)),
		ConstraintOffset: uint32(cs.NbConstraints),
		WireOffset:       uint32(cs.NbInternalVariables + cs.GetNbPublicVariables() + cs.GetNbSecretVariables()),
		BlueprintID:      bID,
	}

	// append the call data
	cs.CallData = append(cs.CallData, calldata...)

	// update the total number of constraints
	blueprint := cs.Blueprints[pi.BlueprintID]
	cs.NbConstraints += blueprint.NbConstraints()

	// add the output wires
	inst := pi.Unpack(cs)
	nbOutputs := blueprint.NbOutputs(inst)
	/***
		Hints: ZhmYe
		blueprint.NbOutputs-> return 0
	***/
	var wires []uint32
	for i := 0; i < nbOutputs; i++ {
		wires = append(wires, uint32(cs.AddInternalVariable()))
	}

	// add the instruction
	cs.Instructions = append(cs.Instructions, pi)

	// update the instruction dependency tree
	iID := len(cs.Instructions) - 1
	// modify by ZhmYe
	blueprint.NewUpdateInstructionTree(inst, cs, iID, cs, true, true)
	//switch Config.Config.Split {
	//case Config.SPLIT_STAGES:
	//	//fmt.Println(cs.GetNbInternalVariables())
	//	blueprint.NewUpdateInstructionTree(inst, cs, iID, cs, true, true)
	//case Config.SPLIT_LEVELS:
	//	level := blueprint.UpdateInstructionTree(inst, cs)
	//	// we can't skip levels, so appending is fine.
	//	if int(level) >= len(cs.Levels) {
	//		cs.Levels = append(cs.Levels, []int{iID})
	//	} else {
	//		cs.Levels[level] = append(cs.Levels[level], iID)
	//	}
	//default:
	//	blueprint.NewUpdateInstructionTree(inst, cs, iID, cs, true, true)
	//}
	//cs.GetDegree(iID)
	return wires
}
func (cs *System) AddInstruction(bID BlueprintID, calldata []uint32) []uint32 {
	// set the offsets
	pi := PackedInstruction{
		StartCallData:    uint64(len(cs.CallData)),
		ConstraintOffset: uint32(cs.NbConstraints),
		WireOffset:       uint32(cs.NbInternalVariables + cs.GetNbPublicVariables() + cs.GetNbSecretVariables()),
		BlueprintID:      bID,
	}

	// append the call data
	cs.CallData = append(cs.CallData, calldata...)

	// update the total number of constraints
	blueprint := cs.Blueprints[pi.BlueprintID]
	cs.NbConstraints += blueprint.NbConstraints()

	// add the output wires
	inst := pi.Unpack(cs)
	nbOutputs := blueprint.NbOutputs(inst)
	/***
		Hints: ZhmYe
		blueprint.NbOutputs-> return 0
	***/
	var wires []uint32
	for i := 0; i < nbOutputs; i++ {
		wires = append(wires, uint32(cs.AddInternalVariable()))
	}

	// add the instruction
	cs.Instructions = append(cs.Instructions, pi)
	// update the instruction dependency tree
	iID := len(cs.Instructions) - 1
	// modify by ZhmYe
	blueprint.NewUpdateInstructionTree(inst, cs, iID, cs, false, false)
	//switch Config.Config.Split {
	//case Config.SPLIT_STAGES:
	//	blueprint.NewUpdateInstructionTree(inst, cs, iID, cs, false, false)
	//case Config.SPLIT_LEVELS:
	//level := blueprint.UpdateInstructionTree(inst, cs)
	//// we can't skip levels, so appending is fine.
	//if int(level) >= len(cs.Levels) {
	//	cs.Levels = append(cs.Levels, []int{iID})
	//} else {
	//	cs.Levels[level] = append(cs.Levels[level], iID)
	//}
	//default:
	//	blueprint.NewUpdateInstructionTree(inst, cs, iID, cs, false, false)
	//}
	//cs.GetDegree(iID)
	return wires
}

// GetNbConstraints returns the number of constraints
func (cs *System) GetNbConstraints() int {
	return cs.NbConstraints
}

func (cs *System) CheckUnconstrainedWires() error {
	// TODO @gbotrel
	return nil
}

func (cs *System) GetR1CIterator() R1CIterator {
	return R1CIterator{cs: cs}
}

func (cs *System) GetSparseR1CIterator() SparseR1CIterator {
	return SparseR1CIterator{cs: cs}
}

func (cs *System) GetCommitments() Commitments {
	return cs.CommitmentInfo
}

// bufPool is a pool of buffers used by getBuffer and putBuffer.
// It is used to avoid allocating buffers for each constraint.
var bufPool = sync.Pool{
	New: func() interface{} {
		r := make([]uint32, 0, 20)
		return &r
	},
}

// getBuffer returns a buffer of at least the given size.
// The buffer is taken from the pool if it is large enough,
// otherwise a new buffer is allocated.
// Caller must call putBuffer when done with the buffer.
func getBuffer() *[]uint32 {
	to := bufPool.Get().(*[]uint32)
	*to = (*to)[:0]
	return to
}

// putBuffer returns a buffer to the pool.
func putBuffer(buf *[]uint32) {
	if buf == nil {
		panic("invalid entry in putBuffer")
	}
	bufPool.Put(buf)
}

func (system *System) AddGkr(gkr GkrInfo) error {
	if system.GkrInfo.Is() {
		return fmt.Errorf("currently only one GKR sub-circuit per SNARK is supported")
	}

	system.GkrInfo = gkr
	return nil
}
