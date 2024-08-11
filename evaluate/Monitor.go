package evaluate

import (
	"github.com/shirou/gopsutil/cpu"
	"runtime"
	"time"
)

// Monitor 用来监控内存和CPU使用情况

type Monitor struct {
	mCPU    bool // 是否监控CPU
	mMemory bool // 是否监控Memory
	//record       MonitorRecord // 监控结果
	finish       bool
	mCPUPercent  float64
	mMemoryUsage float64
}

func NewMonitor(c bool, m bool) Monitor {
	return Monitor{
		mCPU:         c,
		mMemory:      m,
		finish:       false,
		mCPUPercent:  float64(0),
		mMemoryUsage: float64(0),
	}
}

type MonitorRecord struct {
	cpuUsage    float64 // CPU利用率 %
	memoryUsage float64 // 内存使用 GB
}

func (r *MonitorRecord) CPUUsage() float64 {
	return r.cpuUsage
}
func (r *MonitorRecord) MemoryUsage() float64 {
	return r.memoryUsage
}
func (m *Monitor) Start() {
	if m.mCPU {
		go m.startMonitorCPU()
	}
	if m.mMemory {
		go m.startMonitorMemory()
	}
}
func (m *Monitor) Finish() MonitorRecord {
	m.finish = true
	return MonitorRecord{
		cpuUsage:    m.mCPUPercent,
		memoryUsage: m.mMemoryUsage,
	}
}
func (m *Monitor) startMonitorCPU() {
	//mPercent := float64(0)
	for {
		if m.finish {
			break
		}
		percent, _ := cpu.Percent(10*time.Millisecond, false)
		if percent[0] > m.mCPUPercent {
			m.mCPUPercent = percent[0]
		}
	}
}

func (m *Monitor) startMonitorMemory() {
	for {
		if m.finish {
			break
		}
		startTime := time.Now()
		// 这里可以看到整体内存趋势
		//memorySeq := make([]uint64, 0)
		var M runtime.MemStats
		runtime.ReadMemStats(&M)
		startMemory := M.Alloc // 这里也就是PK+CCS
		maxAlloc := startMemory
		//startMemory :
		for {
			//if m.finish {
			//	//fmt.Println(memorySeq)
			//	//r.memory =
			//	r.memory = PackedMemory{
			//		totalMemoryUsed: float64(maxAlloc) / 1024 / 1024 / 1024,
			//		proveMemoryUsed: float64(maxAlloc-startMemory) / 1024 / 1024 / 1024,
			//	}
			//	break
			//}
			if time.Since(startTime) >= time.Duration(10)*time.Millisecond {
				var M runtime.MemStats
				runtime.ReadMemStats(&M)
				nowAlloc := M.Alloc
				if nowAlloc > maxAlloc {
					maxAlloc = nowAlloc
					m.mMemoryUsage = float64(maxAlloc) / 1024 / 1024 / 1024
					//r.memory = float64(maxAlloc) / 1024 / 1024 / 1024
					//r.memory = PackedMemory{
					//	totalMemoryUsed: float64(maxAlloc) / 1024 / 1024 / 1024,
					//	proveMemoryUsed: float64(maxAlloc-startMemory) / 1024 / 1024 / 1024,
					//}
					//fmt.Println(nowAlloc)
				}
				//r.m = append(r.m, nowAlloc)
				//memorySeq = append(memorySeq, nowAlloc)
				//c.m = append(c.m, nowAlloc)
				startTime = time.Now()
			}
		}
	}
}
