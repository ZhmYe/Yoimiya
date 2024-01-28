package graph

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"strconv"
)

type SplitEngine struct {
	forward           *DAG         // 电路Instruction之间的依赖关系(前面的Instruction指向后面的Instruction)组成的DAG
	backward          *DAG         // 电路Instruction之间的依赖关系(后面的Instruction指向前面的Instruction)组成的DAG
	LastLevel         []int        // 记录那些没有后续计算的Instruction
	Stages            []*Stage     // 用于保存Stage, new Stage id = len(Stages)
	RootStages        []*Stage     // 用于保存那些没有父Stage的stage，这些stage从一开始就可以并行运行
	Instruction2Stage map[int]int  // 查询Instruction在哪个Stage中
	stack             []*Stage     // 用于输出
	HasStore          map[int]bool // 用于输出
}

func NewSplitEngine(forward *DAG, backward *DAG, level []int) *SplitEngine {
	s := new(SplitEngine)
	s.forward = forward
	s.backward = backward
	s.LastLevel = level
	s.Stages = make([]*Stage, 0)
	s.RootStages = make([]*Stage, 0)
	s.Instruction2Stage = make(map[int]int)
	s.HasStore = make(map[int]bool)
	return s
}
func (s *SplitEngine) NewStage(Instruction []int) *Stage {
	stage := NewStage(len(s.Stages), Instruction)
	s.Stages = append(s.Stages, stage)
	for _, id := range Instruction {
		s.Instruction2Stage[id] = stage.id
	}
	return stage
}

// checkShuffleDependency 判断是否为宽依赖
func (s *SplitEngine) checkShuffleDependency(id int) bool {
	if !s.forward.Exist(id) {
		fmt.Errorf("id does not exist in DAG")
		return false
	} else {
		return s.forward.SizeOf(id) > 1
	}
}

// checkOneParent 判断当前Instruction的前置依赖是否只有一个
func (s *SplitEngine) checkOneParent(id int) bool {
	if !s.backward.Exist(id) {
		fmt.Errorf("id does not exist in DAG")
		return false
	} else {
		return s.backward.SizeOf(id) == 1
	}
}

// Exist 判断是否Instruction是否已经包含在某个Stage中
func (s *SplitEngine) Exist(id int) bool {
	_, exist := s.Instruction2Stage[id]
	return exist
}
func (s *SplitEngine) getStage(id int) *Stage {
	if !s.Exist(id) {
		fmt.Errorf("Stage does not exist")
		return nil
	}
	return s.Stages[s.Instruction2Stage[id]]
}
func (s *SplitEngine) Split() []*Stage {
	// 遍历所有没有后续计算的Instruction
	for _, iID := range s.LastLevel {
		stage := s.NewStage([]int{iID}) // 新建一个Stage
		s.processStage(iID, stage)
	}
	return s.RootStages
}

type ExamineResult int

const (
	Pass ExamineResult = iota
	RootStageHasParent
	InstructionRepeat
	LinkError
	StageLoss
	StageRepeat
)

func (s *SplitEngine) ClearStack() {
	s.stack = make([]*Stage, 0)
}
func (s *SplitEngine) Examine() ExamineResult {
	// 检验是否有stage被漏记录
	testIdMap := make(map[int]bool)
	for _, stage := range s.Stages {
		if stage.id > len(s.Stages)-1 {
			return StageLoss
		}
		_, exist := testIdMap[stage.id]
		if exist {
			return StageRepeat
		} else {
			testIdMap[stage.id] = true
		}
	}
	for _, stage := range s.RootStages {
		s.appendToStack(stage)
	}
	if len(s.stack) != len(s.Stages) {
		return StageLoss
	}
	s.ClearStack()

	// 检验所有RootStage是否确实没有前置依赖
	for _, stage := range s.RootStages {
		if s.backward.Exist(stage.GetID()) {
			return RootStageHasParent
		}
	}
	// 检验所有Stage的Instruction是否有重复，使用map
	testMap := make(map[int]bool)
	for _, stage := range s.Stages {
		for _, i := range stage.GetInstructions() {
			_, exist := testMap[i]
			if !exist {
				testMap[i] = true
			} else {
				return InstructionRepeat
			}
		}
	}
	// 检验所有stage的父stage个数和实际包含的是否一致
	testLinkMap := make(map[int]int)
	for _, stage := range s.Stages {
		for _, subStage := range stage.GetSubStages() {
			_, exist := testLinkMap[subStage.GetID()]
			if !exist {
				testLinkMap[subStage.GetID()] = 1
			} else {
				testLinkMap[subStage.GetID()] += 1
			}
		}
	}
	for _, stage := range s.Stages {
		if len(stage.GetParentIDs()) != testLinkMap[stage.GetID()] {
			return LinkError
		}
		if testLinkMap[stage.GetID()] != s.backward.SizeOf(stage.GetID()) {
			fmt.Println(testLinkMap[stage.GetID()], s.backward.SizeOf(stage.GetID()))
			return LinkError
		}
	}

	return Pass

}
func (s *SplitEngine) dfs(stage *Stage, testMap map[int]int, number *int) {
	_, exist := testMap[stage.id]
	if exist {
		//fmt.Println("error")
		testMap[stage.id] += 1
		if testMap[stage.id] != len(stage.GetParentIDs()) {
			return
		}
	} else {
		testMap[stage.id] = 1
		if len(stage.GetParentIDs()) > 1 {
			return
		}
	}
	*number++
	for _, sub := range stage.GetSubStages() {
		s.dfs(sub, testMap, number)
	}
}
func (s *SplitEngine) GetRootStages() []*Stage {
	return s.RootStages
}

func (s *SplitEngine) GetStageNumber() int {
	return len(s.Stages)
}
func (s *SplitEngine) processStage(iID int, stage *Stage) {
	if !s.backward.Exist(iID) {
		// 如果当前节点没有父节点
		// 那么当前stage是可以直接运行的，加入到rootStage中
		s.RootStages = append(s.RootStages, stage)
		return
	}
	parents := s.backward.GetLinks(iID) // 获取所有父Instruction iID
	for _, id := range parents {
		// 如果当前连接关系是宽依赖或者当前节点有不止一个父节点
		// 那么针对父节点新建一个Stage，并将当前Stage添加为子Stage
		if s.checkShuffleDependency(id) || !s.checkOneParent(iID) {
			if !s.Exist(id) {
				// 如果还没有父Stage
				ParentStage := s.NewStage([]int{id}) // 新建一个父Stage
				// 添加依赖关系
				ParentStage.AddChild(stage)
				stage.AddParent(ParentStage)
				s.processStage(id, ParentStage)
			} else {
				// 如果有父Stage
				ParentStage := s.getStage(id) // 获取父Stage
				// 添加依赖关系
				ParentStage.AddChild(stage)
				stage.AddParent(ParentStage)
				// 父Stage已经Process过，无需进行processStage
			}
		} else {
			// 当前连接关系是窄依赖并且当前节点只有一个父节点
			// 那么将父节点并入到当前Stage中
			stage.AddInstruction(id, true)
			s.Instruction2Stage[id] = stage.id
			s.processStage(id, stage)
		}
	}
}
func (s *SplitEngine) getTable() table.Writer {
	var (
		colTitleStageID       = "ID"
		colTitleParentID      = "Parent IDs"
		colTitleChildID       = "Child IDs"
		colTitleInstructionID = "Instructions"
		colTitleCount         = "Count"
		colTitleCheck         = "Check(Count = Number of Parents)"
		rowHeader             = table.Row{colTitleStageID, colTitleParentID, colTitleChildID, colTitleInstructionID, colTitleCount, colTitleCheck}
	)
	t := table.NewWriter()
	t.AppendHeader(rowHeader)
	//t.AppendFooter(table.Row{"", "", "Total", 10000})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: colTitleStageID, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
		{Name: colTitleParentID, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
		{Name: colTitleChildID, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
		{Name: colTitleInstructionID, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
		{Name: colTitleCount, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
		{Name: colTitleCheck, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
	})
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateFooter = true
	t.Style().Options.SeparateHeader = true
	t.SetStyle(table.StyleBold)
	//t.SetTitle("Stage Log")
	//fmt.Println(t.Render())
	return t
}
func (s *SplitEngine) appendToStack(stage *Stage) {
	_, exist := s.HasStore[stage.id]
	if exist {
		return
	}
	s.HasStore[stage.id] = true
	s.stack = append(s.stack, stage)
	for _, subStage := range stage.child {
		s.stack = append(s.stack, subStage)
		s.appendToStack(subStage)
	}
}
func shortOutput(output []int) string {
	if len(output) > 5 {
		return strconv.Itoa(output[0]) + " " + strconv.Itoa(output[1]) + " ... " + strconv.Itoa(output[len(output)-2]) + " " + strconv.Itoa(output[len(output)-1])
	}
	if len(output) == 0 {
		return "None"
	}
	str := ""
	for _, element := range output {
		str += strconv.Itoa(element)
		str += " "
	}
	return str[:len(str)-1]
}
func (s *SplitEngine) PrintStages() {
	t := s.getTable()
	for _, stage := range s.RootStages {
		s.appendToStack(stage)
	}
	HasPrint := make(map[int]bool)
	for _, stage := range s.stack {
		_, exist := HasPrint[stage.id]
		if exist {
			continue
		}
		HasPrint[stage.id] = true
		flag := stage.GetCount() == len(stage.GetParentIDs())
		t.AppendRow(table.Row{stage.id, shortOutput(stage.GetParentIDs()), shortOutput(stage.GetChildIDs()), shortOutput(stage.GetInstructions()), stage.count, flag})
	}
	fmt.Println(t.Render())
}
