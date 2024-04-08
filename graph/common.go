package graph

type Layer int

// 用于标记每个stage处在切分后电路的位置，目前只支持二分电路
// TOP表示在上层电路作为中间变量，MIDDLE表示作为上层电路的输出、下层电路的输入，BOTTOM表示在下层电路中作为中间变量
const (
	TOP Layer = iota
	MIDDLE
	BOTTOM
	UNSET
)
