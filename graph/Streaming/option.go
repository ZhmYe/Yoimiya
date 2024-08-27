package Streaming

// PartitionHeuristicsOption Heuristics选项
type PartitionHeuristicsOption int

const (
	Chunking PartitionHeuristicsOption = iota // Divide the stream into chunks of size C and fill the partitions completely in order
	//Balanced                                  // Assign v to a partition of minimal size, breaking ties randomly
)

type PartitionOrderingOption int

const (
	BFS_RANDOM PartitionOrderingOption = iota
	BFS_GREEDY
	DFS_RANDOM
	DFS_GREEDY
)
