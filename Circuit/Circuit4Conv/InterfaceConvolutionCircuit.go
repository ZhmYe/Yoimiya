package Circuit4Conv

import (
	"Yoimiya/Config"
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"Yoimiya/frontend/cs/r1cs"
	"github.com/consensys/gnark-crypto/ecc"
	"time"
)

type InterfaceConvolutionalCircuit struct {
	assignmentGenerator func() frontend.Circuit // 生成assignment
	outerCircuit        frontend.Circuit
	name                string
}

func NewInterfaceConvolutionalCircuit() InterfaceConvolutionalCircuit {
	c := InterfaceConvolutionalCircuit{
		assignmentGenerator: nil,
		outerCircuit:        nil,
		name:                "Conv",
	}
	c.Init()
	return c
}

// GetAssignment 获取测试电路的随机Assignment
func (c *InterfaceConvolutionalCircuit) GetAssignment() frontend.Circuit {
	return c.assignmentGenerator()
}

// Init 初始化
func (c *InterfaceConvolutionalCircuit) Init() {
	var circuit ConvolutionalCircuit
	c.outerCircuit = &circuit
	c.assignmentGenerator = func() frontend.Circuit {
		var A [256][256]frontend.Variable
		for i := 0; i < 256; i++ {
			//A = append(A, make([]frontend.Variable, 100))
			for j := 0; j < 256; j++ {
				A[i][j] = frontend.Variable(0)
			}
		}
		var W [3][3]frontend.Variable
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				W[i][j] = frontend.Variable(0)
			}
		}
		var B [254][254]frontend.Variable
		for i := 0; i < 254; i++ {
			for j := 0; j < 254; j++ {
				B[i][j] = frontend.Variable(0)
			}
		}
		return &ConvolutionalCircuit{A: A, W: W, B: B}
	}
}

func (c *InterfaceConvolutionalCircuit) Compile() (constraint.ConstraintSystem, time.Duration) {
	Config.Config.CancelSplit()
	startTime := time.Now()
	outerCcs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, c.outerCircuit)
	if err != nil {
		panic(err)
	}
	compileTime := time.Since(startTime)
	//fmt.Println("Compile Time:", compileTime)
	return outerCcs, compileTime
}
func (c *InterfaceConvolutionalCircuit) Name() string { return c.name }
