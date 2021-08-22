package internal

import (
	"crontab/worker/pkg"
	"os/exec"
	"time"
)

type Executor struct {}

func InitExecutor() error {
	pkg.G_executor = &Executor{}
	return nil
}

func (e *Executor) ExecuteJob(info *JobExecuteInfo) {
	result := &JobExecuteResult{
		ExecuteInfo: info,
	}

	jobLock := pkg.G_jobMgr.CreateJobLock(info.Job.Name)
	err := jobLock.TryLock()
	if err != nil {
		result.Err = err
		result.EndTime = time.Now()
		pkg.G_scheduler.PushJobResult(result)
		return
	}
	defer jobLock.Unlock()

	result.StartTime = time.Now()
	cmd := exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)
	result.Output, result.Err = cmd.CombinedOutput()
	result.EndTime = time.Now()

	pkg.G_scheduler.PushJobResult(result)
}

