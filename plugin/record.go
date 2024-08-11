package plugin

import (
	"fmt"
	"runtime"
	"time"
)

type PackedTime struct {
	TimeUsed time.Duration
	Name     string
}

func NewPackedTime(name string, t time.Duration) PackedTime {
	return PackedTime{
		TimeUsed: t,
		Name:     name + " Time",
	}
}

type PackedMemory struct {
	TotalMemoryUsed float64 // PK + CCS + Prove
	ProveMemoryUsed float64 // Prove
}
type PluginRecord struct {
	//memory float64 // GB
	Memory PackedMemory
	finish bool
	Times  []PackedTime
	M      []uint64
	Name   string
}

func NewPluginRecord(name string) PluginRecord {
	return PluginRecord{
		Name: name,
		Memory: PackedMemory{
			TotalMemoryUsed: 0,
			ProveMemoryUsed: 0,
		},
		Times:  make([]PackedTime, 0),
		finish: false,
		M:      make([]uint64, 0),
	}
}
func (r *PluginRecord) SetTime(name string, t time.Duration) {
	r.Times = append(r.Times, NewPackedTime(name, t))
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
			r.Memory = PackedMemory{
				TotalMemoryUsed: float64(maxAlloc) / 1024 / 1024 / 1024,
				ProveMemoryUsed: float64(maxAlloc-startMemory) / 1024 / 1024 / 1024,
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
				r.Memory = PackedMemory{
					TotalMemoryUsed: float64(maxAlloc) / 1024 / 1024 / 1024,
					ProveMemoryUsed: float64(maxAlloc-startMemory) / 1024 / 1024 / 1024,
				}
				//fmt.Println(nowAlloc)
			}
			r.M = append(r.M, nowAlloc)
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
	fmt.Printf("[%s Record]: \n", r.Name)
	fmt.Println("\t[Memory Used]: ", r.Memory.TotalMemoryUsed, "GB")
	fmt.Println("\t[Prove Memory Used: ]", r.Memory.ProveMemoryUsed, "GB")
	for _, pt := range r.Times {
		fmt.Printf("\t[%s]: %v \n", pt.Name, pt.TimeUsed)
	}
	//fmt.Println(r.m)
}
