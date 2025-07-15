package scheduler_interface

type Scheduler interface {
	InitScheduler() error
	StartScheduler()
	StopScheduler()
}
