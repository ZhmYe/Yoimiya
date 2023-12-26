package evaluate

const (
	SPLIT_LEVELS = iota
	SPLIT_STAGES
)

type GlobalConfig struct {
	Split int // 对于DAG的划分方式
}

var Config = GlobalConfig{Split: SPLIT_STAGES}
