package plugin

import (
	"fmt"
	"runtime"
	"time"
)

type PluginRecord struct {
	memory      float64 // GB
	proveTime   time.Duration
	compileTime time.Duration
	setupTime   time.Duration
	finish      bool
}

func NewPluginRecord() PluginRecord {
	return PluginRecord{
		memory:      0,
		proveTime:   0,
		compileTime: 0,
		setupTime:   0,
		finish:      false,
	}
}
func (r *PluginRecord) SetCompileTime(t time.Duration) {
	r.compileTime = t
}
func (r *PluginRecord) SetSetupTime(t time.Duration) {
	r.setupTime = t
}
func (r *PluginRecord) SetProveTime(t time.Duration) {
	r.proveTime = t
}
func (r *PluginRecord) MemoryMonitor() {
	maxAlloc := uint64(0)
	startTime := time.Now()
	// 这里可以看到整体内存趋势
	//memorySeq := make([]uint64, 0)
	for {
		if r.finish {
			//fmt.Println(memorySeq)
			r.memory = float64(maxAlloc) / 1024 / 1024 / 1024
			break
		}
		if time.Since(startTime) >= time.Duration(10)*time.Millisecond {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			nowAlloc := m.Alloc
			if nowAlloc > maxAlloc {
				maxAlloc = nowAlloc
				r.memory = float64(maxAlloc) / 1024 / 1024 / 1024
				//fmt.Println(nowAlloc)
			}
			//c.m = append(c.m, nowAlloc)
			startTime = time.Now()
		}
	}
}
func (r *PluginRecord) Finish() {
	r.finish = true
}
func (r *PluginRecord) Print() {
	fmt.Println("[Record]: ")
	fmt.Println("	[Memory Used]: ", r.memory, "GB")
	fmt.Println("	[Compile Time]: ", r.compileTime)
	fmt.Println("	[SetUp Time]: ", r.setupTime)
	fmt.Println("	[Prove Time: ]", r.proveTime)
}
