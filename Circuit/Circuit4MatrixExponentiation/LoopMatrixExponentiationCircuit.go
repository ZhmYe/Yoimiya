package Circuit4MatrixExponentiation

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"github.com/consensys/gnark-crypto/ecc"
	"time"
)

type InterfaceMatrixExponentiationCircuit struct {
	assignmentGenerator func() frontend.Circuit // 生成assignment
	outerCircuit        frontend.Circuit
	name                string
}

func NewInterfaceMatrixExponentiationCircuit() InterfaceMatrixExponentiationCircuit {
	c := InterfaceMatrixExponentiationCircuit{
		assignmentGenerator: nil,
		outerCircuit:        nil,
		name:                "Matrix",
	}
	c.Init()
	return c
}

// GetAssignment 获取测试电路的随机Assignment
func (c *InterfaceMatrixExponentiationCircuit) GetAssignment() frontend.Circuit {
	return c.assignmentGenerator()
}

// Init 初始化
func (c *InterfaceMatrixExponentiationCircuit) Init() {
	var circuit MatrixExponentiationCircuit
	c.outerCircuit = &circuit
	c.assignmentGenerator = func() frontend.Circuit {
		var A [2][2]frontend.Variable
		for i := 0; i < 2; i++ {
			//A = append(A, make([]frontend.Variable, 100))
			for j := 0; j < 2; j++ {
				A[i][j] = frontend.Variable(0)
			}
		}
		return &MatrixExponentiationCircuit{X: A, Y: A}
	}
}

func (c *InterfaceMatrixExponentiationCircuit) Compile() (constraint.ConstraintSystem, time.Duration) {
	startTime := time.Now()
	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, c.outerCircuit)
	if err != nil {
		panic(err)
	}
	compileTime := time.Since(startTime)
	//fmt.Println("Compile Time:", compileTime)
	return outerCcs, compileTime
}
func (c *InterfaceMatrixExponentiationCircuit) Name() string { return c.name }
