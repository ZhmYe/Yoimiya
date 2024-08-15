package Runner

type YoimiyaTaskRunner interface {
	InjectTasks(nbTask int)
	Process()
	Record()
}
