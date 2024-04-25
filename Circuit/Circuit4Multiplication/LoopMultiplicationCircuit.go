package Circuit4Multiplication

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"time"
)

type LoopMultiplicationCircuit struct {
	assignmentGenerator func() frontend.Circuit // 生成assignment
	outerCircuit        frontend.Circuit
}

func NewLoopMultiplicationCircuit() LoopMultiplicationCircuit {
	c := LoopMultiplicationCircuit{
		assignmentGenerator: nil,
		outerCircuit:        nil,
	}
	c.Init()
	return c
}

// GetAssignment 获取测试电路的随机Assignment
func (c *LoopMultiplicationCircuit) GetAssignment() frontend.Circuit {
	return c.assignmentGenerator()
}

// Init 初始化
func (c *LoopMultiplicationCircuit) Init() {
	var circuit MultiplicationCircuit
	c.outerCircuit = &circuit
	c.assignmentGenerator = func() frontend.Circuit {
		return &MultiplicationCircuit{X: 1, Y: 1}
	}
}

func (c *LoopMultiplicationCircuit) Compile() (constraint.ConstraintSystem, time.Duration) {
	startTime := time.Now()
	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, c.outerCircuit)
	if err != nil {
		panic(err)
	}
	compileTime := time.Since(startTime)
	fmt.Println("Compile Time:", compileTime)
	return outerCcs, compileTime
}
