package Sit

// SetPartition 设置某个stage的partition
func (t *SITree) SetPartition(id int, partition int) {
	if id >= len(t.partitions) {
		panic("This stage partition hasn't been set...")
	}
	t.partitions[id] = partition
	// 更新最大的partition
	if partition > t.maxPartition {
		t.maxPartition = partition
	}
}
func (t *SITree) GetPartition(id int) int {
	return t.partitions[id]
}

// InsertWithLayerInNSplit 加上Layer的逻辑
func (t *SITree) InsertWithLayerInNSplit(iID int, previousIds []int) {
	stage := NewStage(-1, iID) // id统一都默认初始化为-1，在append时处理
	// 如果没有父节点
	if len(previousIds) == 0 {
		t.appendStage(stage) // 直接append
		t.appendRoot(stage)  // 没有父节点则一定是root Stage
		// RootStage的Layer标记为Top
		t.SetLayer(stage.GetID(), TOP)
		// rootStage作为第一个partition
		t.SetPartition(stage.GetID(), 1) // 第一个partition

	} else if len(previousIds) == 1 {
		// 如果只有一个父节点
		// 暂时认为当前instruction和父instruction之间是窄依赖关系，合并stage
		previousId := previousIds[0]
		parentStage := t.GetStageByInstruction(previousId) // 得到父stage
		if t.checkSplit(parentStage, previousId) {
			// 需要分裂，那么最终为宽依赖
			fission := t.Split(parentStage, previousId)
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)
			t.appendStage(stage)
			// 判断父节点的layer
			// 这里需要判断父节点的归属情况
			parentPartition := t.GetPartition(parentStage.GetID())
			// 只有一个父节点，那么
			if parentPartition%2 == 0 { // 父节点是作为MIDDLE的
				// 这里父节点可能连接了多个其他partition
				// 父节点保持不变
				// fission得到它的所有子节点里partition最小的
				minPartition := t.maxPartition
				for _, subStage := range fission.GetSubStages() {
					if tmp := t.GetPartition(subStage.GetID()); tmp < minPartition {
						minPartition = tmp
					}
				}
				if minPartition%2 == 0 {
					// 父节点作为middle还连了一个middle，说明两者应该是一样的partition
					if minPartition != parentPartition {
						panic("middle connect another middle!!!")
					}
					t.SetPartition(fission.GetID(), minPartition)
					t.SetPartition(stage.GetID(), minPartition+1)
				} else {
					t.SetPartition(fission.GetID(), minPartition-1)
					// 按照父节点应该放入parentPartition + 1
					// 按照fission应该放入minPartition + 1
					// 这里minPartition >= parentPartition
					if minPartition < parentPartition {
						panic("child's partition should >= parent's")
					}
					t.SetPartition(stage.GetID(), minPartition+1)
				}
			} else if t.maxPartition > parentPartition {
				// 如果父节点的partition比目前最大的partition小，说明它在其所在的partition里是作为TOP的
				// 此时fission节点也划入和父节点一样的partition中
				t.SetPartition(fission.GetID(), parentPartition)
				// 子节点可以作为MIDDLE
				t.SetPartition(stage.GetID(), parentPartition+1) // 用偶数的partition来表示middle
			} else {
				// parentPartition = t.maxPartition
				// 也就是说父节点现在是作为整个sit的bottom的
				// 那么可以新建一个partition，将父节点
			}
			switch t.GetLayer(parentStage.GetID()) {
			case TOP:
				// 父节点依旧保持为TOP
				t.SetLayer(fission.GetID(), TOP) // 将分身节点的layer也设置为TOP,这样无需处理下面的节点变更
				// 将当前stage设置为Middle

				// todo
				t.SetLayer(stage.GetID(), MIDDLE)
				//t.SetLayer(stage.GetID(), TOP) // 如果是单一父节点，不赋值MIDDLE
			case MIDDLE:
				// 如果父节点是Middle，那么将分裂体设置为Bottom，当前stage也设置为Bottom
				// 这里不会影响父节点是否可以作为Middle
				t.SetLayer(fission.GetID(), BOTTOM)
				t.SetLayer(stage.GetID(), BOTTOM)
			case BOTTOM:
				// 如果父节点是Bottom，将分裂体和当前stage设置为Bottom
				t.SetLayer(fission.GetID(), BOTTOM)
				t.SetLayer(stage.GetID(), BOTTOM)
			default:
				panic("Unset Layer Type...")
			}
		} else if len(parentStage.GetSubStages()) != 0 {
			// 如果不需要分裂，但父stage有多个子stage，那么也是宽依赖
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)
			t.appendStage(stage)
			// 如果不需要分裂，父节点有stage
			switch t.GetLayer(parentStage.GetID()) {
			case TOP:
				// 如果父节点是TOP,那么可以将当前节点置为Middle
				// todo
				t.SetLayer(stage.GetID(), MIDDLE) // 只有一个父节点
				//t.SetLayer(stage.GetID(), TOP)
			case MIDDLE:
				// 如果父节点是Middle，此时不会影响父节点作为Middle，因此只需要把当前节点置为BOTTOM
				t.SetLayer(stage.GetID(), BOTTOM)
			case BOTTOM:
				// 如果父节点是BOTTOM，那么显然只需要把当前节点置为BOTTOM
				t.SetLayer(stage.GetID(), BOTTOM)
			default:
				panic("Unset Layer Type...")
			}
		} else {
			// 无需分裂，并且父stage当前没有子stage，那么暂时认为是窄依赖
			t.Combine(stage, parentStage)
			// 如果不需要分裂并且父节点没有stage，那么没有新的stage被添加，因此不需要更改layer
		}
	} else {
		// 如果不止有一个父节点，一定是宽依赖
		hasBeenChild := make(map[int]bool) // 可能会出现多个previousId在同一个stage里面，那么无需后续重新添加child
		// Layer逻辑：
		// 如果不止一个父节点，那么需要遍历所有父节点，此时父节点Layer内容可能变换
		// 尝试把当前节点置为Middle，但需要对所有父节点的Layer情况进行判断
		// 既然当前节点有多个父节点，那么要求其所有一阶父节点均为Middle或均不为Middle
		hasBottom := false
		hasTop := false
		hasMiddle := false
		// 遍历所有父节点
		for _, previousId := range previousIds {
			// 首先针对所有父节点
			parentStage := t.GetStageByInstruction(previousId)
			var fission *Stage
			hasSplit := false
			// 判断是否需要split
			if t.checkSplit(parentStage, previousId) {
				// 需要分裂
				fission = t.Split(parentStage, previousId)
				hasSplit = true
				// 判断父节点类型，对分裂体进行标注，暂时还不对Stage本身进行标注
				switch t.GetLayer(parentStage.GetID()) {
				case TOP:
					// 如果父节点为TOP，那么标注分裂体为TOP
					t.SetLayer(fission.GetID(), TOP)
					hasTop = true
				case BOTTOM:
					// 如果父节点为BOTTOM，那么标注分裂体为BOTTOM
					t.SetLayer(fission.GetID(), BOTTOM)
					hasBottom = true // 出现了Bottom的父节点
				case MIDDLE:
					// 如果父节点为Middle
					// 这里我们令父节点不再为Middle,这样避免处理判断其他父节点是否可以变成Middle的逻辑
					// 将父节点标注为TOP
					t.SetLayer(parentStage.GetID(), TOP)
					// 此时分裂体继承了父节点的所有子节点，并且分裂体有且仅有一个父节点，因此我们可以将分裂体置为Middle
					t.SetLayer(fission.GetID(), MIDDLE)
					//hasMiddle = true
					hasTop = true
				default:
					panic("Unset Layer Type...")
				}
			} else {
				// 如果不需要分裂
				switch t.GetLayer(parentStage.GetID()) {
				case TOP:
					// 如果父节点为TOP
					// 此时不需要做任何处理，需要等待后续处理当前stage的layer
					hasTop = true
				case BOTTOM:
					// 如果父节点为BOTTOM
					hasBottom = true
				case MIDDLE:
					hasMiddle = true
				default:
					panic("Unset Layer Type...")
				}
			}
			_, flag := hasBeenChild[parentStage.GetID()]
			if flag {
				if hasSplit {
					// 分裂的两个stage都是当前stage的父stage
					fission.AddChild(stage)
					stage.AddParent(fission)
					// 处理fission的layer是否需要更改
					switch t.GetLayer(fission.GetID()) {
					case TOP:
						hasTop = true
						// 如果fission是TOP，那么不影响，无需处理
					case BOTTOM:
						// 如果fission是BOTTOM，它也会作为当前节点的父节点
						hasBottom = true
					case MIDDLE:
						// 如果fission是MIDDLE，MIDDLE之间不能相互连接
						// 之前考虑过如果有多个一阶父节点父节点，那么这些一阶父节点父节点要么都为MIDDLE
						// 我们已经将其他非分裂体的父节点全部保证不为MIDDLE
						// 剩下的都是分裂体，且能进入这里的那些分裂体，其父节点也是当前stage的父节点
						// 此时我们可以通过这里的逻辑保证所有这些分裂体都是MIDDLE，因此可以将当前stage置为BOTTOM
						hasMiddle = true
					default:
						panic("Unset Layer Type...")
					}
				}
				continue
			}
			hasBeenChild[parentStage.GetID()] = true
			parentStage.AddChild(stage)
			stage.AddParent(parentStage)

		}
		// 把stage append
		t.appendStage(stage)
		// 遍历完所有的父节点后，根据父节点的类型判断
		if hasTop && !hasMiddle && !hasBottom {
			// 如果只有top的父节点, 那么将当前节点置为middle
			// todo
			//if t.CheckParentSameDepth(stage.GetID()) {
			//	t.SetLayer(stage.GetID(), MIDDLE)
			//} else {
			//	t.SetLayer(stage.GetID(), TOP)
			//}
			t.SetLayer(stage.GetID(), MIDDLE)
		} else if !hasTop && hasMiddle && !hasBottom {
			// 如果只有middle的父节点，那么将当前节点置为bottom
			t.SetLayer(stage.GetID(), BOTTOM)
		} else if !hasTop && !hasMiddle && hasBottom {
			// 如果只有bottom的父节点，那么将当前节点置为bottom
			t.SetLayer(stage.GetID(), BOTTOM)
		} else if !hasTop {
			// 如果没有top的父节点，那么将当前节点置为bottom
			t.SetLayer(stage.GetID(), BOTTOM)
		} else {
			// 如果有top的父节点，那么当前节点置为Bottom，并且递归修改为top的父节点
			t.SetLayer(stage.GetID(), BOTTOM)
			for _, pid := range stage.GetParentIDs() {
				pLayer := t.GetLayer(pid)
				if pLayer == TOP {
					t.switchTop(pid, MIDDLE)
				}
			}
		}
	}
	t.instructions = append(t.instructions, iID)

}
