package SplitPipeline

import (
	groth16 "Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/plugin"
	"runtime"
	"strconv"
	"time"
)

// PipelineConstraintSystem 用于流水线式得到不同的ccs和对应的pk,vk
type PipelineConstraintSystem struct {
	cs constraint.ConstraintSystem
	pk groth16.ProvingKey
	vk groth16.VerifyingKey
	//witness     []int // 记录该电路哪些input是witness
	ibrs        []constraint.IBR
	commitments constraint.Commitments
	coefftable  cs.CoeffTable
	phase       int
	pli         frontend.PackedLeafInfo
}

func NewPipelineConstraintSystem(pli frontend.PackedLeafInfo, ibrs []constraint.IBR, commitment constraint.Commitments, coefftable cs.CoeffTable) (*PipelineConstraintSystem, plugin.PluginRecord) {
	if len(ibrs) == 0 {
		panic("Len(ibrs) == 0")
	}
	record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(0) + " Prepare")
	ibr := ibrs[0]
	runtime.GOMAXPROCS(runtime.NumCPU())
	master := plugin.NewMaster(1)
	//record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(i))
	buildStartTime := time.Now()
	//fmt.Println("	Fill In Sub Circuit ", i, " Witness...")
	//fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), *extras, ecc.BN254.ScalarField())
	//fmt.Println("	Witness Filled...")
	//fmt.Println("	Build Sub Circuit ", i, " From IBR...")
	SubCs, err := plugin.BuildConstraintSystemFromIBR(ibr,
		commitment, coefftable, pli, make([]constraint.ExtraValue, 0), "Sub Circuit"+strconv.Itoa(0))
	//fmt.Println("	Sub Circuit ", i, " Building Finish...")
	if err != nil {
		panic(err)
	}
	record.SetTime("Build", time.Since(buildStartTime))
	//startTime := time.Now()
	pk, vk, setupTime := master.SetUp(SubCs)
	record.SetTime("SetUp", setupTime)
	runtime.GC()
	return &PipelineConstraintSystem{
		cs:          SubCs,
		pk:          pk,
		vk:          vk,
		ibrs:        ibrs,
		commitments: commitment,
		coefftable:  coefftable,
		phase:       0,
		pli:         pli,
	}, record
}
func (cs *PipelineConstraintSystem) Next(record *plugin.PluginRecord) bool {
	if cs.phase == len(cs.ibrs) {
		return false
	}
	cs.phase++
	ibr := cs.ibrs[cs.phase]
	runtime.GOMAXPROCS(runtime.NumCPU())
	master := plugin.NewMaster(1)
	//record := plugin.NewPluginRecord("Sub Circuit " + strconv.Itoa(i))
	buildStartTime := time.Now()
	//fmt.Println("	Fill In Sub Circuit ", i, " Witness...")
	//fullWitness, err := frontend.GenerateSplitWitnessFromPli(pli, ibr.GetWitness(), *extras, ecc.BN254.ScalarField())
	//fmt.Println("	Witness Filled...")
	//fmt.Println("	Build Sub Circuit ", i, " From IBR...")
	SubCs, err := plugin.BuildConstraintSystemFromIBR(ibr,
		cs.commitments, cs.coefftable, cs.pli, make([]constraint.ExtraValue, 0), "Sub Circuit"+strconv.Itoa(cs.phase))
	//fmt.Println("	Sub Circuit ", i, " Building Finish...")
	if err != nil {
		panic(err)
	}
	record.SetTime("Build", time.Since(buildStartTime))
	//startTime := time.Now()
	pk, vk, setupTime := master.SetUp(SubCs)
	record.SetTime("SetUp", setupTime)
	runtime.GC()
	cs.cs, cs.pk, cs.vk = SubCs, pk, vk
	return true
}
func (cs *PipelineConstraintSystem) Params() (constraint.ConstraintSystem, groth16.ProvingKey, groth16.VerifyingKey, []int) {
	return cs.cs, cs.pk, cs.vk, cs.ibrs[cs.phase].GetWitness()
}
