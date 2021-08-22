package pkg

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cron_expr"`
}

type JobEvent struct {
	EventType int
	Job       *Job
}
