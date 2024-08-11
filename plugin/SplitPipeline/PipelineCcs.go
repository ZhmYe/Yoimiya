package SplitPipeline

import (
	"Yoimiya/Config"
	groth16 "Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"runtime"
	"strconv"
)

// PipelineConstraintSystem 用于流水线式得到不同的ccs和对应的pk,vk
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
	//master := plugin.NewMaster(1)
	////record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(i))
	//buildStartTime := time.Now()
	////fmt.Println("	Fill In Sub Circuit ", i, " Witness...")
	////fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), *extras, ecc.BN254.ScalarField())
	////fmt.Println("	Witness Filled...")
	////fmt.Println("	Build Sub Circuit ", i, " From IBR...")
	//SubCs, err := plugin.BuildConstraintSystemFromIBR(ibr,
	//	commitment, coefftable, pli, make([]constraint.ExtraValue, 0), "Sub Circuit"+strconv.Itoa(0))
	////fmt.Println("	Sub Circuit ", i, " Building Finish...")
	//if err != nil {
	//	panic(err)
	//}
	//record.SetTime("Build", time.Since(buildStartTime))
	////startTime := time.Now()
	//pk, vk, setupTime := master.SetUp(SubCs)
	//record.SetTime("SetUp", setupTime)
	//runtime.GC()
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
		pcs.SetUp()
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

//func (cs *PipelineConstraintSystem) Next(record *plugin.PluginRecord) bool {
//	if cs.phase == len(cs.ibrs) {
//		return false
//	}
//	cs.phase++
//	ibr := cs.ibrs[cs.phase]
//	//newExtra :=
//	runtime.GOMAXPROCS(runtime.NumCPU())
//	master := plugin.NewMaster(1)
//	//record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(i))
//	buildStartTime := time.Now()
//	//fmt.Println("	Fill In Sub Circuit ", i, " Witness...")
//	//fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), *extras, ecc.BN254.ScalarField())
//	//fmt.Println("	Witness Filled...")
//	//fmt.Println("	Build Sub Circuit ", i, " From IBR...")
//	SubCs, err := plugin.BuildConstraintSystemFromIBR(ibr,
//		cs.commitments, cs.coefftable, cs.pli, cs.extra, "Sub Circuit"+strconv.Itoa(cs.phase))
//	//fmt.Println("	Sub Circuit ", i, " Building Finish...")
//	if err != nil {
//		panic(err)
//	}
//	record.SetTime("Sub Circuit "+strconv.Itoa(cs.phase)+"Build", time.Since(buildStartTime))
//	//startTime := time.Now()
//	pk, vk, setupTime := master.SetUp(SubCs)
//	record.SetTime("Sub Circuit "+strconv.Itoa(cs.phase)+"Setup", setupTime)
//	runtime.GC()
//	cs.cs, cs.pk, cs.vk = SubCs, pk, vk
//	newExtra := SubCs.(*cs_bn254.R1CS).GetForwardOutputs()
//	cs.extra = append(cs.extra, newExtra...)
//	return true
//}

func (cs *PipelineConstraintSystem) Params() (constraint.ConstraintSystem, groth16.ProvingKey, groth16.VerifyingKey, []int) {
	return cs.pcs[cs.phase].Params()
}
