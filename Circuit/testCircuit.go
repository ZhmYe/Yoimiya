package Circuit

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"time"
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

// 这里写一个testCircuit的接口
// 方便代码编写
type TestCircuit interface {
	GetAssignment() frontend.Circuit                       // 获取测试电路的随机Assignment
	Compile() (constraint.ConstraintSystem, time.Duration) // 编译电路得到ccs,并返回编译时间
	Name() string
	//TestMisalignedParalleling(nbTask int, cut int) Record  // 测试错位并行效果
	//TestNSplit(n int) Record                               // 测试split效果
	//TestSerialRunning(nbTask int) Record                   // 测试正常串行证明n个电路的效果
}
