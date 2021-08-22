package internal

import (
	"context"
	"crontab/pkg"
	pkg2 "crontab/worker/pkg"
	"github.com/gorhill/cronexpr"
	"time"
)

type JobSchedulePlan struct {
	Job      *pkg.Job
	Expr     *cronexpr.Expression
	NextTime time.Time
}

type JobExecuteInfo struct {
	Job      *pkg.Job
	PlanTime time.Time
	RealTime time.Time
	CancelCtx context.Context
	CancelFunc context.CancelFunc
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output      []byte
	Err         error
	StartTime   time.Time
	EndTime     time.Time
}

type Scheduler struct {
	JobEventChan chan *pkg.JobEvent
	JobPlanTable map[string]*JobSchedulePlan
	JobExecutingTable map[string]*JobExecuteInfo
	JobResultChan chan *JobExecuteResult
}

func InitScheduler() {
	pkg2.G_scheduler = &Scheduler{
		JobEventChan: make(chan *pkg.JobEvent, 1000),
		JobPlanTable: make(map[string]*JobSchedulePlan),
		JobExecutingTable: make(map[string]*JobExecuteInfo),
		JobResultChan: make(chan *JobExecuteResult, 1000),
	}
}

func (s *Scheduler) ScheduleLoop() {
	timer := time.NewTimer(s.TrySchedule())

	for {
		select {
			case jobEvent := <- s.JobEventChan:
				s.handleJobEvent(jobEvent)
			case result := <- s.JobResultChan:
				s.handleJobResult(result)
			case <-timer.C:
		}

		timer.Reset(s.TrySchedule())
	}
}

func (s *Scheduler) handleJobEvent(jobEvent *pkg.JobEvent)  {
	switch jobEvent.EventType {
		case pkg.ADD:
			plan, err := BuildJobSchedulePlan(jobEvent.Job)
			if err != nil {
				return
			}
			s.JobPlanTable[jobEvent.Job.Name] = plan
		case pkg.DELETE:
			if _, exists := s.JobPlanTable[jobEvent.Job.Name]; exists {
				delete(s.JobPlanTable, jobEvent.Job.Name)
			}
		case pkg.KILL:
			if info, exists := s.JobExecutingTable[jobEvent.Job.Name]; exists  {
				info.CancelFunc()
			}
	}

}

func (s *Scheduler) TrySchedule() time.Duration {
	if len(s.JobPlanTable) == 0 {
		return 1 * time.Second
	}

	var nearTime *time.Time
	now := time.Now()
	for _, plan := range s.JobPlanTable {
		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			s.TryStartJob(plan)
			plan.NextTime = plan.Expr.Next(now)
		}

		if nearTime == nil || plan.NextTime.Before(*nearTime) {
			nearTime = &plan.NextTime
		}
	}

	return nearTime.Sub(time.Now())
}

func (s *Scheduler) TryStartJob(jobPlan *JobSchedulePlan) {
	if _, executing := s.JobExecutingTable[jobPlan.Job.Name]; !executing {
		executeInfo := &JobExecuteInfo{
			Job:      jobPlan.Job,
			PlanTime: jobPlan.NextTime,
			RealTime: time.Now(),
		}
		executeInfo.CancelCtx, executeInfo.CancelFunc = context.WithCancel(context.Background())
		s.JobExecutingTable[jobPlan.Job.Name] = executeInfo

		go pkg2.G_executor.ExecuteJob(executeInfo)
	}
}

func (s *Scheduler) handleJobResult(result *JobExecuteResult)  {
	delete(s.JobExecutingTable, result.ExecuteInfo.Job.Name)
	log := pkg2.BuildJobLog(result)
	pkg2.G_logSink.Append(log)
}


func BuildJobSchedulePlan(job *pkg.Job) (*JobSchedulePlan, error) {
	parse, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return nil, err
	}

	jobPlan := &JobSchedulePlan{
		Job: job,
		Expr: parse,
		NextTime: parse.Next(time.Now()),
	}

	return jobPlan, nil
}

func (s *Scheduler) PushJobResult(jobResult *JobExecuteResult)  {
	select {
		case s.JobResultChan <- jobResult:
		default:
	}
}