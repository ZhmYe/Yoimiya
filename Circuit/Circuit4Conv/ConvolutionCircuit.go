package Circuit4Conv

import "Yoimiya/frontend"

// ConvolutionalCircuit 卷积，卷积核为私有输入, f(A,w) = B
// 这里简单定padding = 0,stride = 1
type ConvolutionalCircuit struct {
	A [256][256]frontend.Variable `gnark:",public"`
	W [3][3]frontend.Variable     `gnark:",filter"`
	B [254][254]frontend.Variable `gnark:"C"`
	//C frontend.Variable `gnark:"C"`
}

func (c *ConvolutionalCircuit) Define(api frontend.API) error {
	//sum := frontend.Variable(0)
	var tmp [254][254]frontend.Variable
	// todo 计算卷积逻辑
	for row := 0; row < 254; row++ {
		for col := 0; col < 254; col++ {
			// (row, col)就是卷积核左上角的位置
			// 两个矩阵对应位置相乘相加
			tmp[row][col] = frontend.Variable(0)
			for k := 0; k < 9; k++ {
				tmp[row][col] = api.Add(tmp[row][col], api.Mul(c.A[row+k/3][col+k%3], c.W[k/3][k%3]))
			}
		}
	}
	for i := 0; i < 254; i++ {
		for j := 0; j < 254; j++ {
			api.AssertIsEqual(c.B[i][j], tmp[i][j])
		}
	}
	return nil
}
