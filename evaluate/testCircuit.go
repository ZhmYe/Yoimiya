package evaluate

import (
	"Yoimiya/constraint"
	"Yoimiya/frontend"
	"time"
)

// 这里写一个testCircuit的接口
// 方便代码编写
type testCircuit interface {
	GetAssignment() frontend.Circuit                       // 获取测试电路的随机Assignment
	Compile() (constraint.ConstraintSystem, time.Duration) // 编译电路得到ccs,并返回编译时间
	Name() string
	//TestMisalignedParalleling(nbTask int, cut int) Record  // 测试错位并行效果
	//TestNSplit(n int) Record                               // 测试split效果
	//TestSerialRunning(nbTask int) Record                   // 测试正常串行证明n个电路的效果
}
