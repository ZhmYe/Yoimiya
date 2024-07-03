package LRO_Tree

type LroNode interface {
	Depth() int   // 返回深度，Input Node Depth = 0
	IsRoot() bool // 是否为根节点
	NotRoot()
	AddDegree()
	SetSplit(s int) bool
	Ergodic(b *Bucket)
	Degree() int // 返回出度
	TryVisit() bool
	CheckMiddle(s int, b *Bucket)
	IsMiddle() bool
}
