package Circuit4Conv

import "Yoimiya/frontend"

// ConvolutionalCircuit 卷积，卷积核为私有输入, f(A,w) = B
// 这里简单定padding = 0,stride = 1
type ConvolutionalCircuit struct {
	A [200][200]frontend.Variable `gnark:",public"`
	W [3][3]frontend.Variable     `gnark:",filter"`
	B [198][198]frontend.Variable `gnark:"C"`
	//C frontend.Variable `gnark:"C"`
}

func (c *ConvolutionalCircuit) Define(api frontend.API) error {
	//sum := frontend.Variable(0)
	var tmp [198][198]frontend.Variable
	// todo 计算卷积逻辑
	for row := 0; row < 198; row++ {
		for col := 0; col < 198; col++ {
			// (row, col)就是卷积核左上角的位置
			// 两个矩阵对应位置相乘相加
			tmp[row][col] = frontend.Variable(0)
			for k := 0; k < 9; k++ {
				tmp[row][col] = api.Add(tmp[row][col], api.Mul(c.A[row+k/3][col+k%3], c.W[k/3][k%3]))
			}
		}
	}
	for i := 0; i < 198; i++ {
		for j := 0; j < 198; j++ {
			api.AssertIsEqual(c.B[i][j], tmp[i][j])
		}
	}
	return nil
}
