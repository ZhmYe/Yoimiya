package constraint

import (
	"S-gnark/debug"
	"math"
)

type Level int

const (
	LevelUnset Level = -1
)

type InstructionTree interface {
	// InsertWire inserts a wire in the instruction tree at the given level.
	// If the wire is already in the instruction tree, it panics.
	InsertWire(wire uint32, level Level)

	// HasWire returns true if the wire is in the instruction tree.
	// False if it's a constant or an input.
	HasWire(wire uint32) bool

	// GetWireLevel returns the level of the wire in the instruction tree.
	// If HasWire(wire) returns false, behavior is undefined.
	GetWireLevel(wire uint32) Level
	IsInputOrConstant(wire uint32, split bool) bool
}

// the instruction tree is a simple array of levels.
// it's morally a map[uint32 (wireID)]PackedLevel, but we use an array for performance reasons.

func (system *System) HasWire(wireID uint32) bool {
	offset := system.internalWireOffset()
	if wireID < offset {
		// it's an input.
		return false
	}
	// if wireID == maxUint32, it's a constant.
	//fmt.Println(len(system.lbWireLevel), system.NbInternalVariables)
	//return (wireID - offset) < uint32(system.NbInternalVariables)
	return (wireID - offset) < uint32(system.NbInternalVariables) // modify by ZhmYe, to delete lbWireLevel
}

// IsInputOrConstant add by ZhmYe
func (system *System) IsInputOrConstant(wireID uint32, split bool) bool {
	return !system.IsOutput(wireID, split)
}

// IsOutput 这里如果是第n次切割电路,wireId会溢出
// todo 已经在外部导入了input(public、private)
// 所以还是可以按照offset判断是否为input
func (system *System) IsOutput(wireID uint32, split bool) bool {
	offset := system.internalWireOffset()
	if wireID == math.MaxUint32 {
		// const
		return false
	}
	//bias := system.GetWireBias(int(wireID)) // 得到wireID在values中的位置
	bias, exist := system.Bias[wireID]
	if !split {
		// 如果不是在split方法下
		// 此时bias未设置
		if wireID < offset {
			return false
		}
	} else {
		// 如果是在split方法下
		// 如果bias被设置，那么确认bias是否比offset小
		if exist {
			if bias < int(offset) {
				// it's an input.
				return false
			}
			// 如果bias比offset大，说明是output
			return true
		} else {
			// 如果bias未被设置，说明该结果不是input
			return true
		}
	}
	//bias, exist := system.Bias[wireID]
	//if exist {
	//	return bias-int(offset) < system.NbInternalVariables
	//}
	// if wireID == maxUint32, it's a constant.
	//fmt.Println(len(system.lbWireLevel), system.NbInternalVariables)
	//return (wireID - offset) < uint32(system.NbInternalVariables)
	//bias := system.GetWireBias(int(wireID))
	return true
	//return (wireID-offset) < uint32(system.NbInternalVariables) || !split // modify by ZhmYe, to delete lbWir
}
func (system *System) GetWireLevel(wireID uint32) Level {
	return system.lbWireLevel[wireID-system.internalWireOffset()]
}

func (system *System) InsertWire(wireID uint32, level Level) {
	if debug.Debug {
		if level < 0 {
			panic("level must be >= 0")
		}
		if wireID < system.internalWireOffset() {
			panic("cannot insert input wire in instruction tree")
		}
	}
	wireID -= system.internalWireOffset()
	if system.lbWireLevel[wireID] != LevelUnset {
		panic("wire already exist in instruction tree")
	}

	system.lbWireLevel[wireID] = level
}

// internalWireOffset returns the position of the first internal wire in the wireIDs.
func (system *System) internalWireOffset() uint32 {
	return uint32(system.GetNbPublicVariables() + system.GetNbSecretVariables())
}
