package scheduler

import (
    "context"
    "fmt"
    "github.com/coreos/etcd/clientv3"
    "github.com/kackerx/crontab/pkg/common"
)

var GLock *Lock

type Lock struct {
    Kv         clientv3.KV
    Lease      clientv3.Lease
    LeaseID    clientv3.LeaseID
    CancelFunc context.CancelFunc
    IsLocked   bool
    LockKey    string
}

func NewLock(jobName string, client *clientv3.Client) *Lock {
    return &Lock{
        Kv:      client.KV,
        Lease:   client.Lease,
        LockKey: jobName,
    }
}

func (lock *Lock) TryLock() error {
    var (
        leaseGrantResp *clientv3.LeaseGrantResponse
        cancelCtx      context.Context
        cancelFunc     context.CancelFunc
        keepaliveCh    <-chan *clientv3.LeaseKeepAliveResponse
        txn            clientv3.Txn
        txnResp        *clientv3.TxnResponse
        leaseId        clientv3.LeaseID
        err            error
    )

    leaseGrantResp, err = lock.Lease.Grant(context.TODO(), 5)
    if err != nil {
        goto FAIL
    }

    leaseId = leaseGrantResp.ID

    // 自动续租, cncelctx用于取消自动续租
    cancelCtx, cancelFunc = context.WithCancel(context.TODO())
    keepaliveCh, err = lock.Lease.KeepAlive(cancelCtx, leaseId)
    if err != nil {
        goto FAIL
    }

    // * 处理续租应答
    go func() {
        for {
            select {
            case keepResp := <-keepaliveCh:
                if keepResp == nil {
                    goto END
                }
            }
        }
    END:
    }()

    txn = lock.Kv.Txn(context.TODO()).
        If(clientv3.Compare(clientv3.CreateRevision(common.JOB_LOCK_DIR+lock.LockKey), "=", 0)).
        Then(clientv3.OpPut(common.JOB_LOCK_DIR+lock.LockKey, "", clientv3.WithLease(leaseId))).
        Else(clientv3.OpGet(common.JOB_LOCK_DIR + lock.LockKey))

    txnResp, err = txn.Commit()
    if err != nil {
        fmt.Println(err)
        goto FAIL
    }

    if !txnResp.Succeeded {
        // 抢锁失败
        err = common.ERR_LOCKED
        goto FAIL
    }

    // 抢锁成功
    lock.LeaseID = leaseId
    lock.CancelFunc = cancelFunc
    lock.IsLocked = true
    return nil

FAIL:
    cancelFunc() // 取消自动续租
    lock.Lease.Revoke(context.TODO(), leaseId)
    return err
}

func (lock *Lock) UnLock() {
    if lock.IsLocked {
        lock.IsLocked = false
        lock.CancelFunc()
        lock.Lease.Revoke(context.TODO(), lock.LeaseID)
    }
}
