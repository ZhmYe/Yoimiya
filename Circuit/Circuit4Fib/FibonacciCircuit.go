package Circuit4Fib

import "C"
import (
	"Yoimiya/Config"
	"Yoimiya/frontend"
)

// FibonacciCircuit 斐波那契数列，循环计算
type FibonacciCircuit struct {
	X1 frontend.Variable `gnark:",public"` // a_1,a_2
	X2 frontend.Variable `gnark:",public"`
	V1 frontend.Variable `gnark:"v1"` // an-1
	V2 frontend.Variable `gnark:"v2"` // an
	A  frontend.Variable `gnark:",public"`
	B  frontend.Variable `gnark:",public"`
}

// 这里简单起见，测试时X1,X2初始化为0，0，不然数字太大
// 这里如果按照斐波那契数列正常的通项，加法不会产生约束
// 简单的乘法，如果有常数会作为复制约束而不是产生约束
// 因此这里只能将通项改成了a_n^2 + a_{n+1}^2

func (c *FibonacciCircuit) Define(api frontend.API) error {
	for i := 0; i < Config.Config.NbLoop; i++ {
		c.X1 = api.Add(api.Mul(c.A, c.X1), api.Mul(c.B, c.X2))
		c.X2 = api.Add(api.Mul(c.A, c.X1), api.Mul(c.B, c.X2))
	}
	api.AssertIsEqual(c.V1, c.X1)
	api.AssertIsEqual(c.V2, c.X2)
	return nil
}
