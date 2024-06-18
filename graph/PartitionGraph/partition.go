package PartitionGraph

const PID_UNSET int = -1

// Partition 一个分区
// 每个分区包含一些节点，每个节点之间的拓扑序组织为Level
// 可以根据levels来分裂Partition

type Partition struct {
	levels [][]int
	//index  map[int]int // node->level
	pID    int          // PartitionID
	isRoot bool         // 是否为根分区
	child  []*Partition // 子分区
}

func NewPartition() *Partition {
	p := Partition{
		levels: make([][]int, 0),
		//index:  make(map[int]int),
		pID:    PID_UNSET,
		isRoot: false,
		child:  make([]*Partition, 0),
	}
	return &p
}
func (p *Partition) IsRoot() bool {
	return p.isRoot
}
func (p *Partition) AssignRoot() {
	p.isRoot = true
}
func (p *Partition) AddLevel(level []int) {
	p.levels = append(p.levels, level)
}

// Insert 重构levels，判断是否需要分裂
// 返回值包括两种情况
// 如果需要分裂，那么后面的[]*Partition就是新的分裂分区，里面的所有Node需要在外面重新赋予PID和Level
// 如果不需要分裂，则返回插入的节点所在的Level(int)
func (p *Partition) Insert(iID int, previousLevels []int) (bool, int, []*Partition) {
	maxLevel := -1
	for _, level := range previousLevels {
		//level, exist := p.index[id]
		//if !exist {
		//	panic("This Partition don't have this instruction!!!")
		//}
		if level > maxLevel {
			maxLevel = level
		}
	}
	maxLevel++
	if maxLevel >= len(p.levels) {
		// 新增一个level，无需分裂
		p.levels = append(p.levels, []int{iID})
	} else {
		// 也就是这一层level，出现了两个可并行的内容，这里实现分裂
		p.levels[maxLevel] = append(p.levels[maxLevel], iID)

		return true, -1, p.handleSplit(maxLevel)
	}
	return false, maxLevel, make([]*Partition, 0)
}

// handleSplit 分裂，返回分裂后的子Partition
// 分裂函数传入level字段，说明该level出现了两个可并行的内容
// 将原本的该level(:-1)和其下所有level划分为一个新的partition，并将level[-1]本身作为一个新的partition
// todo 这里是不是可以对split的粒度进行调整？
func (p *Partition) handleSplit(level int) []*Partition {
	// 首先考虑如果分裂的位置在第一层level,那么分裂以后应该只有两个分区：原分区和一个包括新节点的分区
	// 如果分裂的位置在第一层level，那么就是说插入的节点没有父节点
	// 除了根分区，insert的时候都是根据父节点所在分区来插入的，说明只要是调用Insert一定有父节点，那么不可能在第一层
	// 因此如果分裂的位置在第一层，一定就是根分区的根节点插入，那么就简单的将每次根节点的插入
	remainLevel := make([][]int, 0)
	for i := 0; i < level; i++ {
		remainLevel = append(remainLevel, p.levels[i])
	}
	leftPartition := NewPartition()
	leftPartition.AddLevel(p.levels[level][:len(p.levels)-1])
	for i := level + 1; i < len(p.levels); i++ {
		leftPartition.AddLevel(p.levels[i])
	}
	rightPartition := NewPartition()
	rightPartition.AddLevel([]int{p.levels[level][len(p.levels)-1]})
	p.levels = remainLevel
	// 添加子分区
	p.AddChild(leftPartition, rightPartition)
	return []*Partition{leftPartition, rightPartition}
}

func (p *Partition) SetPID(id int) {
	p.pID = id
}
func (p *Partition) AddChild(partitions ...*Partition) {
	for _, partition := range partitions {
		p.child = append(p.child, partition)
	}
}
