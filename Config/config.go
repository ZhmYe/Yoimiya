package Config

type SPLIT_METHOD int

const (
	SPLIT_LEVELS SPLIT_METHOD = iota
	SPLIT_STAGES
)

type CODE_MODE int

const (
	DEBUG CODE_MODE = iota
	NORMAL
)

type SPLIT_MODE int

const (
	CLUSTER SPLIT_MODE = iota
	ALONE
)

type GlobalConfig struct {
	Split SPLIT_METHOD // 对于DAG的划分方式
	//MaxParallelingNumber int          // 最大并发数
	MinWorkPerCPU int
	Mode          CODE_MODE
	SplitMode     SPLIT_MODE
	rootPath      string
}

var Config = GlobalConfig{
	Split: SPLIT_LEVELS,
	//MaxParallelingNumber: 100,
	MinWorkPerCPU: 50,
	Mode:          NORMAL,
	SplitMode:     ALONE,
	rootPath:      "/root/Yoimiya/logWriter/log/",
}

func (c *GlobalConfig) IsCluster() bool {
	if c.SplitMode == CLUSTER {
		return true
	}
	return false
}
