package pkg

import "crontab/worker/internal"

type JobLog struct {
	JobName      string `bson:"jobName"`
	Command      string `bson:"command"`
	Err          string `bson:"err"`
	Output       string `bson:"output"`
	PlanTime     int64  `bson:"planTime"`
	ScheduleTime int64  `bson:"scheduleTime"`
	StartTime    int64  `bson:"startTime"`
	EndTime      int64  `bson:"endTime"`
}

func BuildJobLog(result *internal.JobExecuteResult) *JobLog {
	log := &JobLog{
		JobName:      result.ExecuteInfo.Job.Name,
		Command:      result.ExecuteInfo.Job.Command,
		Output:       string(result.Output),
		PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
		ScheduleTime: result.ExecuteInfo.RealTime.Unix() / 1000 / 1000,
		StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
		EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
	}
	if result.Err != nil {
		log.Err = result.Err.Error()
	} else {
		log.Err = ""
	}
	return log
}
