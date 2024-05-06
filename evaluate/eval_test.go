package evaluate

import (
	"Yoimiya/Circuit/Circuit4Multiplication"
	_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	_ "net/http/pprof"                // 必须，引入 pprof 模块
	"runtime"
	"testing"
)

// MisAlignedParalleling测试
// 2278000304 40984537466
// 1778632328 60993852813

// split测试
func TestMisalignedParalleling(t *testing.T) {
	//circuit := Circuit4VerifyCircuit.NewVerifyCircuit()
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	circuit := Circuit4Multiplication.NewLoopMultiplicationCircuit()
	instance := Instance{circuit: &circuit}
	//record := instance.TestSerialRunning(4)
	//fmt.Println(record)
	runtime.GOMAXPROCS(runtime.NumCPU())
	record := instance.TestMisalignedParalleling(4, 2)
	record.Sprintf(true, "test_misaligned_paralleling")
}

func TestMemoryReduceByNSplit(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	circuit := Circuit4Multiplication.NewLoopMultiplicationCircuit()
	instance := Instance{circuit: &circuit}
	record := instance.TestNSplit(2)
	record.Sprintf(true, "test_n_split")
	record = instance.TestNormal()
	record.Sprintf(true, "test_normal_running")
}
