package Circuit4Fib

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"time"
)

type LoopFibonacciCircuit struct {
	assignmentGenerator func() frontend.Circuit // 生成assignment
	outerCircuit        frontend.Circuit
}

func NewLoopFibonacciCircuit() LoopFibonacciCircuit {
	c := LoopFibonacciCircuit{
		assignmentGenerator: nil,
		outerCircuit:        nil,
	}
	c.Init()
	return c
}

// GetAssignment 获取测试电路的随机Assignment
func (c *LoopFibonacciCircuit) GetAssignment() frontend.Circuit {
	return c.assignmentGenerator()
}

// Init 初始化
func (c *LoopFibonacciCircuit) Init() {
	var circuit FibonacciCircuit
	c.outerCircuit = &circuit
	c.assignmentGenerator = func() frontend.Circuit {
		return &FibonacciCircuit{X1: 0, X2: 0, V1: 0, V2: 0}
	}
}

func (c *LoopFibonacciCircuit) Compile() (constraint.ConstraintSystem, time.Duration) {
	startTime := time.Now()
	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, c.outerCircuit)
	if err != nil {
		panic(err)
	}
	compileTime := time.Since(startTime)
	fmt.Println("Compile Time:", compileTime)
	return outerCcs, compileTime
}