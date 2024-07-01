package evaluate

import (
	"Yoimiya/Circuit/Circuit4Conv"
	"Yoimiya/Circuit/Circuit4Fib"
	"Yoimiya/Circuit/Circuit4MatrixMultiplication"
	"Yoimiya/Circuit/Circuit4Multiplication"
	"Yoimiya/Circuit/Circuit4VerifyCircuit"
)

type CircuitOption int

// 这里给出所有电路的枚举
const (
	Fib       CircuitOption = iota // 这个需要注释compress或者修改config.CompressThreshold才能得到
	FibSquare                      // 带平方的变种斐波那契数列电路
	Mul                            // 连乘电路
	Verify                         // 递归验证电路
	Matrix                         // 矩阵乘法
	Conv                           // 卷积
)

// format 用于format log的名字
func format(circuitName string, testName string) string {
	return "test_" + testName + "_" + circuitName
}

// getCircuit 根据电路枚举给出电路
func getCircuit(option CircuitOption) testCircuit {
	switch option {
	case Fib:
		circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		return &circuit
	case FibSquare:
		circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		return &circuit
	case Mul:
		circuit := Circuit4Multiplication.NewLoopMultiplicationCircuit()
		return &circuit
	case Verify:
		circuit := Circuit4VerifyCircuit.NewVerifyCircuit()
		return &circuit
	case Matrix:
		circuit := Circuit4MatrixMultiplication.NewInterfaceMatrixMultiplicationCircuit()
		return &circuit
	case Conv:
		circuit := Circuit4Conv.NewInterfaceConvolutionalCircuit()
		return &circuit
	default:
		circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		return &circuit
	}
}
