package pkg

import (
	"crontab/worker/configs"
	"crontab/worker/internal"
)

var (
	G_config    *configs.Config
	G_jobMgr    *JobMgr
	G_scheduler *internal.Scheduler
	G_executor  *internal.Executor
	G_logSink   *LogSink
)
