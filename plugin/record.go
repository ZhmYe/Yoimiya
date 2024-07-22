package plugin

import (
	"fmt"
	"runtime"
	"time"
)

type PackedTime struct {
	timeUsed time.Duration
	name     string
}

func NewPackedTime(name string, t time.Duration) PackedTime {
	return PackedTime{
		timeUsed: t,
		name:     name + " Time",
	}
}

type PackedMemory struct {
	totalMemoryUsed float64 // PK + CCS + Prove
	proveMemoryUsed float64 // Prove
}
type PluginRecord struct {
	//memory float64 // GB
	memory PackedMemory
	finish bool
	times  []PackedTime
	m      []uint64
	name   string
}

func NewPluginRecord(name string) PluginRecord {
	return PluginRecord{
		name: name,
		memory: PackedMemory{
			totalMemoryUsed: 0,
			proveMemoryUsed: 0,
		},
		times:  make([]PackedTime, 0),
		finish: false,
		m:      make([]uint64, 0),
	}
}
func (r *PluginRecord) SetTime(name string, t time.Duration) {
	r.times = append(r.times, NewPackedTime(name, t))
}

func (r *PluginRecord) MemoryMonitor() {
	//maxAlloc := uint64(0)
	startTime := time.Now()
	// 这里可以看到整体内存趋势
	//memorySeq := make([]uint64, 0)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startMemory := m.Alloc // 这里也就是PK+CCS
	maxAlloc := startMemory
	//startMemory :
	for {
		if r.finish {
			//fmt.Println(memorySeq)
			//r.memory =
			r.memory = PackedMemory{
				totalMemoryUsed: float64(maxAlloc) / 1024 / 1024 / 1024,
				proveMemoryUsed: float64(maxAlloc-startMemory) / 1024 / 1024 / 1024,
			}
			break
		}
		if time.Since(startTime) >= time.Duration(10)*time.Millisecond {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			nowAlloc := m.Alloc
			if nowAlloc > maxAlloc {
				maxAlloc = nowAlloc
				//r.memory = float64(maxAlloc) / 1024 / 1024 / 1024
				r.memory = PackedMemory{
					totalMemoryUsed: float64(maxAlloc) / 1024 / 1024 / 1024,
					proveMemoryUsed: float64(maxAlloc-startMemory) / 1024 / 1024 / 1024,
				}
				//fmt.Println(nowAlloc)
			}
			r.m = append(r.m, nowAlloc)
			//memorySeq = append(memorySeq, nowAlloc)
			//c.m = append(c.m, nowAlloc)
			startTime = time.Now()
		}
	}
}
func (r *PluginRecord) Finish() {
	r.finish = true
}
func (r *PluginRecord) Print() {
	fmt.Printf("[%s Record]: \n", r.name)
	fmt.Println("\t[Memory Used]: ", r.memory.totalMemoryUsed, "GB")
	fmt.Println("\t[Prove Memory Used: ]", r.memory.proveMemoryUsed, "GB")
	for _, pt := range r.times {
		fmt.Printf("\t[%s]: %v \n", pt.name, pt.timeUsed)
	}
	//fmt.Println(r.m)
}
