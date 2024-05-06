package Circuit4Fib

import "Yoimiya/frontend"

// FibonacciCircuit 斐波那契数列，循环计算
type FibonacciCircuit struct {
	//x frontend.Variable
	X1, X2 frontend.Variable `gnark:"public"` // a_1,a_2
	////N      frontend.Variable `gnark:"public"` // 循环的次数
	V1 frontend.Variable `gnark:"v1"` // an-1
	V2 frontend.Variable `gnark:"v2"` // an
	//X frontend.Variable `gnark:"public"`
	//Y frontend.Variable `gnark:"y"`
}

func (c *FibonacciCircuit) Define(api frontend.API) error {
	var fib [1000000]frontend.Variable // 记录fib的整个序列
	fib[0] = api.Add(c.X1, 0)
	fib[1] = api.Add(c.X2, 0)
	for i := 2; i < 1000000; i++ {
		fib[i] = api.Add(fib[i-2], fib[i-1])
	}
	api.AssertIsEqual(c.V1, fib[len(fib)-2])
	api.AssertIsEqual(c.V2, fib[len(fib)-1])
	return nil
}
