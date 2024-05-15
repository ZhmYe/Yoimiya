package Record

import (
	"Yoimiya/logWriter"
	"fmt"
	"strconv"
	"time"
)

// PackedTime 可能有很多时间需要设置
type PackedTime struct {
	//startTime time.Time // 记录起始时间
	CompileTime time.Duration // 编译时间
	//RunTime     time.Duration // 总运行时间
	SplitTime time.Duration // split花费的时间，现在split中构建sit或level的过程本身就包含在compile里
	SetUpTime time.Duration // SetUp花费的时间
	SolveTime time.Duration // Solve + run所花费的时间
	BuildTime time.Duration // 构建上下半电路的时间
}

func NewPackedTime() *PackedTime {
	return &PackedTime{
		//startTime: time.Now(),
		CompileTime: 0,
		//RunTime:     0,
		SplitTime: 0,
		SetUpTime: 0,
		SolveTime: 0,
		BuildTime: 0,
	}
}
func (t *PackedTime) SetCompileTime(compileTime time.Duration) {
	t.CompileTime = t.CompileTime + compileTime
}

func (t *PackedTime) SetSplitTime(splitTime time.Duration) {
	t.SplitTime = t.SplitTime + splitTime
}
func (t *PackedTime) SetSetUpTime(setUptime time.Duration) {
	t.SetUpTime = t.SetUpTime + setUptime
}
func (t *PackedTime) SetSolveTime(solveTime time.Duration) {
	t.SolveTime = t.SolveTime + solveTime
}
func (t *PackedTime) SetBuildTime(buildTime time.Duration) {
	t.BuildTime = t.BuildTime + buildTime
}
func (t *PackedTime) ClearTime() {
	t.SolveTime = 0
	t.SetUpTime = 0
	//t.RunTime = 0
	t.CompileTime = 0
	t.SplitTime = 0
	t.BuildTime = 0
}

// Record 记录运行时的一些数据
type Record struct {
	PackedTime               // 时间
	memory     int           // 程序使用多少内存
	slotTime   time.Duration // misaligned测试总运行时间
}

func (r *Record) SetMemory(m int) {
	r.memory = m
}

func (r *Record) SetSlotTime(slotTime time.Duration) {
	r.slotTime = slotTime
}
func (r *Record) Clear() {
	r.ClearTime()
	r.memory = 0
	r.slotTime = 0
}
func (r *Record) Sprintf(log bool, dir string, path string) {
	if log {
		r.Log(dir, path)
		return
	}
	fmt.Println("[Record]: ")
	fmt.Println("	[Memory Used]: ", float32(r.memory)/1024/1024/1024, "GB")
	fmt.Println("	[Total Time]: ", r.slotTime)
	fmt.Println("	[Detailed Time]: ")
	fmt.Println("		[Compile Time]: ", r.PackedTime.CompileTime)
	fmt.Println("		[Set Up Time]: ", r.PackedTime.SetUpTime)
	fmt.Println("		[Solve Time]: ", r.PackedTime.SolveTime)
	fmt.Println("		[Split Time]: ", r.PackedTime.SplitTime)
	fmt.Println("		[Build Time]: ", r.PackedTime.BuildTime)
}
func (r *Record) Log(dir string, path string) {
	lw := logWriter.NewLogWriter(dir + "/" + "record_log_" + path)
	lw.Writeln("[Record]: ")
	lw.Writeln("	[Memory Used]: " + strconv.FormatFloat(float64(r.memory)/1024/1024/1024, 'f', -1, 32) + "GB")
	lw.Writeln("	[Total Time]: " + r.slotTime.String())
	lw.Writeln("	[Detailed Time]: ")
	lw.Writeln("		[Compile Time]: " + r.PackedTime.CompileTime.String())
	lw.Writeln("		[Set Up Time]: " + r.PackedTime.SetUpTime.String())
	lw.Writeln("		[Solve Time]: " + r.PackedTime.SolveTime.String())
	lw.Writeln("		[Split Time]: " + r.PackedTime.SplitTime.String())
	lw.Writeln("		[Build Time]: " + r.PackedTime.BuildTime.String())
	lw.Finish()
}
func NewRecord() *Record {
	return &Record{
		PackedTime: *NewPackedTime(),
		memory:     0,
		slotTime:   0,
	}
}

var GlobalRecord = NewRecord()
