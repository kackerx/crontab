package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	if err != nil {
		panic(err)
	}

	lock := Lock{
		Kv:         client.KV,
		Lease:      clientv3.NewLease(client),
		LeaseID:    0,
		CancelFunc: nil,
		IsLocked:   false,
	}

	for i := 0; i < 3; i++ {
		go func() {
			for {
				if err := lock.TryLock(); err != nil {
					fmt.Println(err)
					time.Sleep(time.Second)
					continue
				}
				fmt.Println("执行业务逻辑")
				time.Sleep(time.Second * 2)
				lock.UnLock()
				break
			}
		}()
	}

	time.Sleep(time.Second * 100)
}

type Lock struct {
	Kv         clientv3.KV
	Lease      clientv3.Lease
	LeaseID    clientv3.LeaseID
	CancelFunc context.CancelFunc
	IsLocked   bool
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
		If(clientv3.Compare(clientv3.CreateRevision("/cron/kill/test"), "=", 0)).
		Then(clientv3.OpPut("/cron/kill/test", "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/kill/test"))

	txnResp, err = txn.Commit()
	if err != nil {
		fmt.Println(err)
		goto FAIL
	}

	if !txnResp.Succeeded {
		// 抢锁失败
		err = errors.New("lock fail")
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
	fmt.Println(lock.IsLocked)
	if lock.IsLocked {
		fmt.Println("释放锁")
		lock.IsLocked = false
		lock.CancelFunc()
		lock.Lease.Revoke(context.TODO(), lock.LeaseID)
	}
}
