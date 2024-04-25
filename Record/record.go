package Record

import "time"

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

//func (t *PackedTime) SetRunTime(runTime time.Duration) {
//	t.RunTime = t.RunTime + runTime
//}

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
func NewRecord() *Record {
	return &Record{
		PackedTime: *NewPackedTime(),
		memory:     0,
		slotTime:   0,
	}
}

var GlobalRecord = NewRecord()
