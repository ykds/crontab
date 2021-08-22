package pkg

import (
	"context"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobLock struct {
	kv    clientv3.KV
	lease clientv3.Lease

	LockName   string
	cancelFunc context.CancelFunc
	leaseId    clientv3.LeaseID
	isLocked   bool
}

func InitJobLock(lockName string, kv clientv3.KV, lease clientv3.Lease) *JobLock {
	return &JobLock{
		kv:       kv,
		lease:    lease,
		LockName: lockName,
	}
}

func (l *JobLock) TryLock() error {
	grant, err := l.lease.Grant(context.Background(), 5)
	if err != nil {
		return err
	}

	cancel, cancelFunc := context.WithCancel(context.Background())
	_, err = l.lease.KeepAlive(cancel, grant.ID)
	if err != nil {
		cancelFunc()
		_, _ = l.lease.Revoke(context.Background(), grant.ID)
		return err
	}

	txn := l.kv.Txn(context.Background())
	txn.If(clientv3.Compare(clientv3.CreateRevision(l.LockName), "=", 0)).Then(clientv3.OpPut(l.LockName, "", clientv3.WithLease(grant.ID))).Else(clientv3.OpGet(l.LockName))
	commit, err := txn.Commit()
	if err != nil {
		cancelFunc()
		_, _ = l.lease.Revoke(context.Background(), grant.ID)
		return err
	}
	if !commit.Succeeded {
		cancelFunc()
		_, _ = l.lease.Revoke(context.Background(), grant.ID)
		return errors.New("lock fail")
	}

	l.cancelFunc = cancelFunc
	l.leaseId = grant.ID
	l.isLocked = true
	return nil
}

func (l *JobLock) Unlock() {
	if l.isLocked {
		l.cancelFunc()
		_, _ = l.lease.Revoke(context.Background(), l.leaseId)
	}
}
