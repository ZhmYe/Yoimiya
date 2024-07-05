package plugin

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"time"
)

// todo 先还是按照只有电路间并行的写

// Slot 一个子电路
type Slot struct {
	pk     groth16.ProvingKey
	vk     groth16.VerifyingKey
	pcs    PackedConstraintSystem
	id     int
	buffer *Buffer
}
type SlotReceipt struct {
	sID       int // slot id
	tID       int // task id
	proof     split.PackedProof
	extra     []constraint.ExtraValue
	proveTime time.Duration
}

func (s *Slot) Process() SlotReceipt {
	if s.IsEmpty() {
		panic("Slot is Empty!!!")
	}
	return s.Prove(s.buffer.Pop())
	//receipts := make([]SlotReceipt, len(s.buffer.items))
	////for !s.buffer.IsEmpty() {
	//var wg sync.WaitGroup
	//wg.Add(len(s.buffer.items))
	//for i := 0; i < len(s.buffer.items); i++ {
	//	tmp := i
	//	go func(i int) {
	//		receipts[i] = s.Prove(s.buffer.items[i])
	//		wg.Done()
	//	}(tmp)
	//}
	//wg.Wait()
	////}
	//return receipts
}
func (s *Slot) Setup() {
	pk, vk, err := s.pcs.SetUp()
	if err != nil {
		panic(err)
	}
	s.pk, s.vk = pk, vk
	runtime.GC()
}
func (s *Slot) Prove(request BParam) SlotReceipt {
	//fmt.Println("Start Prove request")
	startTime := time.Now()
	cs := s.pcs.CS()
	proof, err := groth16.Prove(cs, s.pk.(groth16.ProvingKey), request.witness)
	//Record.GlobalRecord.SetSolveTime(time.Since(startTime))
	if err != nil {
		panic(err)
	}
	proveTime := time.Since(startTime)
	//var nextExtra []fr.Element
	//nextForwardOutput, nextExtra := GetExtra(split)
	GetExtra := func(system constraint.ConstraintSystem) []constraint.ExtraValue {
		switch _r1cs := system.(type) {
		case *cs_bn254.R1CS:
			return _r1cs.GetForwardOutputs()
		default:
			panic("Only Support bn254 r1cs now...")
		}
	}
	newExtra := GetExtra(cs)
	//*extra = make([]any, 0)
	//for _, e := range newExtra {
	//	newE := constraint.NewExtraValue(e.GetWireID())
	//	newE.SetValue(e.GetValue())
	//	*extra = append(*extra, newE)
	//}
	publicWitness, _ := request.witness.Public()
	runtime.GC()
	return SlotReceipt{
		sID:       s.id,
		tID:       request.tID,
		proof:     split.NewPackedProof(proof, s.vk, publicWitness),
		extra:     newExtra,
		proveTime: proveTime,
	}
}
func (s *Slot) HandleRequest(tID int, pli frontend.PackedLeafInfo, extra []constraint.ExtraValue) {
	witness, err := frontend.GenerateSplitWitnessFromPli(pli, s.pcs.Witness(), extra, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)
	}
	s.buffer.Push(tID, witness, extra)
}
func (s *Slot) IsEmpty() bool {
	return s.buffer.IsEmpty()
}
