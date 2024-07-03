package Config

type SPLIT_METHOD int

const (
	SPLIT_LEVELS SPLIT_METHOD = iota
	SPLIT_STAGES
	SPLIT_LRO
)

type GlobalConfig struct {
	Split SPLIT_METHOD // 对于DAG的划分方式
	//MaxParallelingNumber int          // 最大并发数
	MinWorkPerCPU     int
	isSplit           bool
	RootPath          string
	CompressThreshold int
	NbLoop            int
}

var Config = GlobalConfig{
	Split: SPLIT_LRO,
	//MaxParallelingNumber: 100,
	MinWorkPerCPU:     50,
	RootPath:          "/root/Yoimiya/logWriter/log/",
	CompressThreshold: 500,
	NbLoop:            1000000,
	isSplit:           true,
}

func (c *GlobalConfig) IsSplit() bool {
	return c.isSplit
}
func (c *GlobalConfig) SwitchToSplit() {
	c.isSplit = true
	c.Split = SPLIT_LRO

}
func (c *GlobalConfig) CancelSplit() {
	c.isSplit = false
	c.Split = SPLIT_LEVELS
}
