package factory

import (
	"event_service/internal/executor"
	"event_service/internal/executor/executor_interface"
)

func ExecutorFactory(name string) executor_interface.Executor {
	switch name {
	case "default":
		return executor.NewDefaultExecutor()
	case "mail":
		return executor.NewMailExecutor()
	}
	return nil
}
