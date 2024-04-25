package evaluate

import (
	"Yoimiya/Circuit/Circuit4VerifyCircuit"
	"fmt"
	"testing"
)

func TestMisalignedParalleling(t *testing.T) {
	circuit := Circuit4VerifyCircuit.NewVerifyCircuit()
	//loopMultiplicationCircuit := Circuit4Multiplication.NewLoopMultiplicationCircuit()
	instance := Instance{circuit: &circuit}
	record := instance.TestSerialRunning(2)
	fmt.Println(record)
	record = instance.TestMisalignedParalleling(2, 2)
	fmt.Println(record)
}
