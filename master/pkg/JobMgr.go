package pkg

import (
	"context"
	"crontab/pkg"
	"encoding/json"
	"github.com/pkg/errors"
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

// SaveJob 保存任务到 etcd
func (j *JobMgr) SaveJob(namespace string, job *pkg.Job) (*pkg.Job, error) {
	if !strings.HasSuffix(namespace, "/") {
		namespace = namespace + "/"
	}

	key := namespace + job.Name
	content, err := json.Marshal(job)
	if err != nil {
		return nil, errors.Wrapf(err, "解析json异常")
	}

	putResponse, err := j.kv.Put(context.Background(), key, string(content), clientv3.WithPrevKV())
	if err != nil {
		return nil, errors.Wrapf(err, "put job fail")
	}

	if putResponse.PrevKv != nil {
		preJob := &pkg.Job{}
		err = json.Unmarshal(putResponse.PrevKv.Value, preJob)
		if err != nil {
			return nil, errors.Wrapf(err, "解析pre job异常")
		}
		return preJob, nil
	}
	return nil, nil
}

// DeleteJob 从 etcd 删除任务
func (j *JobMgr) DeleteJob(namespace, jobName string) (*pkg.Job, error) {
	if !strings.HasSuffix(namespace, "/") {
		namespace = namespace + "/"
	}

	key := namespace + jobName
	response, err := j.kv.Delete(context.Background(), key, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	if response.PrevKvs != nil && len(response.PrevKvs) > 0 {
		value := response.PrevKvs[0].Value
		job := &pkg.Job{}
		err := json.Unmarshal(value, job)
		if err == nil {
			return job, nil
		}
	}
	return nil, nil
}


// ListJobs 从 etcd 查询任务列表
func (j *JobMgr) ListJobs(namespace string) ([]*pkg.Job, error) {
	response, err := j.kv.Get(context.Background(), namespace, clientv3.WithPrefix())
	if err != nil {
		return nil, err
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

	return jobList, nil
}

// KillJob 杀死正在运行的任务
func (j *JobMgr) KillJob(jobName string) error {

	key := "/cron/kill/" + jobName

	// 创建一个 1 秒的租约，只要把需要杀死的任务添加到 etcd，etcd 通知对应 jobEvent 监听器后就可以删除该键，无需长时间保存
	response, err := j.lease.Grant(context.Background(), 1)
	if err != nil {
		return err
	}

	_, err = j.kv.Put(context.Background(), key, "", clientv3.WithLease(response.ID))
	if err != nil {
		return err
	}
	return nil
}
