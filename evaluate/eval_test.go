package evaluate

import (
	"Yoimiya/Circuit/Circuit4Multiplication"
	"fmt"
	_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	"net/http"
	_ "net/http/pprof" // 必须，引入 pprof 模块
	"runtime"
	"testing"
)

func TestMisalignedParalleling(t *testing.T) {
	go func() {
		// terminal: $ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
		// web:
		// 1、http://localhost:8081/ui
		// 2、http://localhost:6060/debug/charts
		// 3、http://localhost:6060/debug/pprof
		fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	//circuit := Circuit4VerifyCircuit.NewVerifyCircuit()
	//runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	circuit := Circuit4Multiplication.NewLoopMultiplicationCircuit()
	instance := Instance{circuit: &circuit}
	//record := instance.TestSerialRunning(4)
	//fmt.Println(record)
	runtime.GOMAXPROCS(runtime.NumCPU())
	record := instance.TestMisalignedParalleling(4, 2)
	fmt.Println(record)
}
