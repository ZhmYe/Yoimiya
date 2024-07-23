package SplitPipeline

type Groth16Runner interface {
	InjectTasks(nbTask int)
	Process()
	Record()
}
