package Yoimiya

func (y *Yoimiya) Run() {
	switch y.mode {
	case Serial:
		y.SerialRunningImpl()
	case Pipeline:
	case SplitPipeline:
	default:

	}
}
func (y *Yoimiya) SerialRunningImpl() {

}
