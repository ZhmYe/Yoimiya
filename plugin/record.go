package plugin

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
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
type PackedCPUUsage struct {
	AverageCPUUsage float64 // 平均
	MaxCPUUsage     float64
	track           []float64
}
type PluginRecord struct {
	//memory float64 // GB
	Memory PackedMemory
	//cpuUsage
	cpuUsage PackedCPUUsage
	finish   bool
	Times    []PackedTime
	M        []uint64
	Name     string
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

func (r *PluginRecord) CPUUsageMonitor() {
	//startTime := time.Now()
	track := make([]float64, 0)
	total := float64(0)
	maxU := float64(0)
	for {
		if r.finish {
			//fmt.Println(memorySeq)
			//r.memory =
			r.cpuUsage = PackedCPUUsage{
				AverageCPUUsage: total / float64(len(track)),
				MaxCPUUsage:     maxU,
				track:           track,
			}
			break
		}
		cmd := exec.Command("bash", "-c", "mpstat 1 1 | awk '/^[0-9]/ {print $3}'")
		output, err := cmd.Output()
		if err != nil {
			//panic(err)
			continue
		}
		//fmt.Println(string(output))
		// 只保留第二行（实际的%usr数值），去掉空白字符
		lines := strings.Split(string(output), "\n")
		//fmt.Println(len(lines))
		if len(lines) > 1 {
			result := strings.TrimSpace(lines[1])
			//fmt.Println(result)
			percent, err := strconv.ParseFloat(result, 64)
			if err != nil {
				//fmt.Println("Error converting to float:", err)
				//return
				//panic(err)
				continue
			}
			//fmt.Println(percent)
			if percent > maxU {
				maxU = percent
			}
			track = append(track, percent)
			total += percent
			r.cpuUsage = PackedCPUUsage{
				AverageCPUUsage: total / float64(len(track)),
				MaxCPUUsage:     maxU,
				track:           track,
			}
			//fmt.Println(percent)
		}
	}
}
func (r *PluginRecord) MemoryMonitor() {
	//maxAlloc := uint64(0)
	//cmd := exec.Command("sh", "/root/Yoimiya/cpu_usage_test.sh")
	//output, err := cmd.Output()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(111)
	//fmt.Println(string(output))
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
	//fmt.Println(r.M)
	fmt.Printf("[%s Record]: \n", r.Name)
	fmt.Println("\t[Memory Used]: ", r.Memory.TotalMemoryUsed, "GB")
	fmt.Println("\t[Prove Memory Used: ]", r.Memory.ProveMemoryUsed, "GB")
	fmt.Println("\t[Max CPU Usage]: ", r.cpuUsage.MaxCPUUsage, "%")
	fmt.Println("\t [Average CPU Usage]: ", r.cpuUsage.AverageCPUUsage)
	fmt.Println(r.cpuUsage.track)
	for _, pt := range r.Times {
		fmt.Printf("\t[%s]: %v \n", pt.Name, pt.TimeUsed)
	}
	//fmt.Println(r.m)
}
