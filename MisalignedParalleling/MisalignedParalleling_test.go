package MisalignedParalleling

import (
	"Yoimiya/constraint"
	"Yoimiya/evaluate/Circuit4VerifyCircuit"
	"Yoimiya/frontend"
	"testing"
)

func Test4MisalignedParalleling(t *testing.T) {
	master := NewMParallelingMaster()
	p1, p2, p3, p4 := Circuit4VerifyCircuit.GetVerifyCircuitParam()
	assignmentGenerator := func() frontend.Circuit {
		assignment, _ := Circuit4VerifyCircuit.GetVerifyCircuitAssignment(p1, p2, p3, p4)
		return assignment
	}
	csGenerator := func() constraint.ConstraintSystem {
		_, outerCircuit := Circuit4VerifyCircuit.GetVerifyCircuitAssignment(p1, p2, p3, p4)
		return Circuit4VerifyCircuit.GetVerifyCircuitCs(outerCircuit)
	}
	master.Initialize(2, 2, csGenerator, assignmentGenerator)
	master.Start()
}
