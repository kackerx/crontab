package global

import (
	"github.com/kackerx/crontab/internal/scheduler"
)

var (
	G_executor  *scheduler.Executor
	G_scheduler *scheduler.Scheduler
)
