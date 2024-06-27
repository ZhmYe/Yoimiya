package Circuit4MatrixMultiplication

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"time"
)

type InterfaceMatrixMultiplicationCircuit struct {
	assignmentGenerator func() frontend.Circuit // 生成assignment
	outerCircuit        frontend.Circuit
	name                string
}

func NewInterfaceMatrixMultiplicationCircuit() InterfaceMatrixMultiplicationCircuit {
	c := InterfaceMatrixMultiplicationCircuit{
		assignmentGenerator: nil,
		outerCircuit:        nil,
		name:                "loop_multiplication",
	}
	c.Init()
	return c
}

// GetAssignment 获取测试电路的随机Assignment
func (c *InterfaceMatrixMultiplicationCircuit) GetAssignment() frontend.Circuit {
	return c.assignmentGenerator()
}

// Init 初始化
func (c *InterfaceMatrixMultiplicationCircuit) Init() {
	var circuit MatrixMultiplicationCircuit
	c.outerCircuit = &circuit
	c.assignmentGenerator = func() frontend.Circuit {
		var A [150][150]frontend.Variable
		for i := 0; i < 150; i++ {
			//A = append(A, make([]frontend.Variable, 100))
			for j := 0; j < 150; j++ {
				A[i][j] = frontend.Variable(0)
			}
		}
		return &MatrixMultiplicationCircuit{A: A, B: A, C: A}
	}
}

func (c *InterfaceMatrixMultiplicationCircuit) Compile() (constraint.ConstraintSystem, time.Duration) {
	startTime := time.Now()
	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, c.outerCircuit)
	if err != nil {
		panic(err)
	}
	compileTime := time.Since(startTime)
	fmt.Println("Compile Time:", compileTime)
	return outerCcs, compileTime
}
func (c *InterfaceMatrixMultiplicationCircuit) Name() string { return c.name }
