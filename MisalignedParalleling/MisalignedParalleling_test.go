package MisalignedParalleling

import (
	Circuit4VerifyCircuit2 "Yoimiya/Circuit/Circuit4VerifyCircuit"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"testing"
)

func Test4MisalignedParalleling(t *testing.T) {
	master := NewMParallelingMaster()
	p1, p2, p3, p4 := Circuit4VerifyCircuit2.GetVerifyCircuitParam()
	assignmentGenerator := func() frontend.Circuit {
		assignment, _ := Circuit4VerifyCircuit2.GetVerifyCircuitAssignment(p1, p2, p3, p4)
		return assignment
	}
	csGenerator := func() constraint.ConstraintSystem {
		_, outerCircuit := Circuit4VerifyCircuit2.GetVerifyCircuitAssignment(p1, p2, p3, p4)
		return Circuit4VerifyCircuit2.GetVerifyCircuitCs(outerCircuit)
	}
	master.Initialize(2, 2, csGenerator, assignmentGenerator)
	master.Start()
}
