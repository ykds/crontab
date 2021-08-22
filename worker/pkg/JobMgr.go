package pkg

import (
	"context"
	"crontab/pkg"
	"encoding/json"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

func InitJobMgr(c *pkg.EtcdConfig) (*JobMgr, error) {
	config := clientv3.Config{
		Endpoints:   c.EtcdEndpoints,
		DialTimeout: time.Duration(c.EtcdDialTimeout) * time.Second,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	jobMgr := &JobMgr{
		client:  client,
		kv:      clientv3.NewKV(client),
		lease:   clientv3.NewLease(client),
		watcher: clientv3.NewWatcher(client),
	}
	return jobMgr, nil
}

func (j *JobMgr) ListJobs(namespace string) ([]*pkg.Job, int64, error) {
	response, err := j.kv.Get(context.Background(), namespace, clientv3.WithPrefix())
	if err != nil {
		return nil, 0, err
	}

	jobList := make([]*pkg.Job, 0)
	if response.Kvs != nil {
		for _, value := range response.Kvs {
			job := &pkg.Job{}
			err := json.Unmarshal(value.Value, job)
			if err != nil {
				continue
			}
			jobList = append(jobList, job)
		}
	}

	return jobList, response.Header.Revision, nil
}

func (j *JobMgr) WatchJobs() error {
	jobs, revision, err := j.ListJobs("/cron/jobs/")
	if err != nil {
		return err
	}

	var jobEvent *pkg.JobEvent

	for _, job := range jobs {
		jobEvent = &pkg.JobEvent{
			EventType: pkg.ADD,
			Job:       job,
		}
		G_scheduler.JobEventChan <- jobEvent
	}

	// listen job change
	go func() {
		watchChan := j.watcher.Watch(context.Background(), "/cron/jobs/", clientv3.WithRev(revision+1), clientv3.WithPrefix())

		for resp := range watchChan {
			for _, event := range resp.Events {
				switch event.Type {
				case mvccpb.PUT:
					job, err2 := unpackJob(event.Kv.Value)
					if err2 != nil {
						continue
					}

					jobEvent = &pkg.JobEvent{
						EventType: pkg.ADD,
						Job:       job,
					}

				case mvccpb.DELETE:
					name := extractJobName(string(event.Kv.Key))

					jobEvent = &pkg.JobEvent{
						EventType: pkg.DELETE,
						Job:       &pkg.Job{Name: name},
					}
				}
				G_scheduler.JobEventChan <- jobEvent
			}
		}
	}()

	return nil
}

func (j *JobMgr) WatchKillJob() {
	watch := j.client.Watch(context.Background(), "/cron/kill/", clientv3.WithPrefix())

	for resp := range watch {
		for _, event := range resp.Events {
			switch event.Type {
			case mvccpb.PUT:
				jobName := strings.TrimPrefix(string(event.Kv.Key), "/cron/kill/")
				jobEvent := &pkg.JobEvent{
					EventType: pkg.KILL,
					Job:       &pkg.Job{Name: jobName},
				}
				G_scheduler.JobEventChan <- jobEvent
			}
		}
	}
}

func (j *JobMgr) CreateJobLock(jobName string) *pkg.JobLock {
	return pkg.InitJobLock("/cron/lock/"+jobName, j.kv, j.lease)
}

func extractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, "/cron/jobs/")
}

func unpackJob(value []byte) (*pkg.Job, error) {
	job := &pkg.Job{}
	err := json.Unmarshal(value, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}
