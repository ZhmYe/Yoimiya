package SplitPipeline

import (
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"runtime"
	"strconv"
	"time"
)

// PipelineConstraintSystem 用于流水线式得到不同的ccs和对应的witnessID
type PipelineConstraintSystem struct {
	//cs constraint.ConstraintSystem
	//pk groth16.ProvingKey
	//vk groth16.VerifyingKey
	//witness     []int // 记录该电路哪些input是witness
	//ibrs        []constraint.IBR
	pcs         []PackedConstraintSystem
	commitments constraint.Commitments
	coefftable  cs_bn254.CoeffTable
	phase       int
	pli         frontend.PackedLeafInfo
	extra       []constraint.ExtraValue
}

func NewPipelineConstraintSystem(pli frontend.PackedLeafInfo, ibrs []constraint.IBR, commitment constraint.Commitments, coefftable cs_bn254.CoeffTable) *PipelineConstraintSystem {
	if len(ibrs) == 0 {
		panic("Len(ibrs) == 0")
	}
	//record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(0) + " Prepare")
	//ibr := ibrs[0]
	runtime.GOMAXPROCS(runtime.NumCPU())
	pcs := PipelineConstraintSystem{
		//cs:          SubCs,
		//pk:          pk,
		//vk:          vk,
		//ibrs:        ibrs,
		pcs:         make([]PackedConstraintSystem, 0),
		commitments: commitment,
		coefftable:  coefftable,
		phase:       -1,
		pli:         pli,
		extra:       make([]constraint.ExtraValue, 0),
	}
	pcs.buildPackedConstraintSystems(pli, ibrs, make([]constraint.ExtraValue, 0))
	return &pcs
}
func (cs *PipelineConstraintSystem) buildPackedConstraintSystems(pli frontend.PackedLeafInfo, ibrs []constraint.IBR, extras []constraint.ExtraValue) {
	Config.Config.CancelSplit()
	for i, ibr := range ibrs {
		pcs, newExtras := BuildNewPackedConstraintSystem(pli, ibr, extras)
		extras = append(extras, newExtras...)
		pcs.SetCommitment(cs.commitments)
		pcs.SetCoeff(cs.coefftable)
		plugin.PrintConstraintSystemInfo(pcs.CS().(*cs_bn254.R1CS), "Sub Circuit "+strconv.Itoa(i))
		//pcs.SetUp()
		cs.pcs = append(cs.pcs, pcs)
		runtime.GC()
	}
}
func (cs *PipelineConstraintSystem) Next() bool {
	if cs.phase == len(cs.pcs)-1 {
		return false
	}
	cs.phase++
	return true
}

func (cs *PipelineConstraintSystem) Params() (constraint.ConstraintSystem, []int) {
	return cs.pcs[cs.phase].Params()
}
func (cs *PipelineConstraintSystem) GetParams(phase int) (constraint.ConstraintSystem, []int) {
	return cs.pcs[phase].Params()
}
func (cs *PipelineConstraintSystem) SetUp() ([]groth16.ProvingKey, []groth16.VerifyingKey, time.Duration) {
	pks := make([]groth16.ProvingKey, 0)
	vks := make([]groth16.VerifyingKey, 0)
	startTime := time.Now()
	for _, packedCcs := range cs.pcs {
		ccs, _ := packedCcs.Params()
		pk, vk, err := groth16.Setup(ccs)
		if err != nil {
			panic(err)
		}
		pks = append(pks, pk)
		vks = append(vks, vk)
	}
	return pks, vks, time.Since(startTime)
}
func (cs *PipelineConstraintSystem) Len() int {
	return len(cs.pcs)
}
