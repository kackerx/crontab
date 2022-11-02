package service

import (
    "context"
    "fmt"
    "github.com/coreos/etcd/clientv3"
    "github.com/coreos/etcd/mvcc/mvccpb"
    "github.com/kackerx/crontab/internal/scheduler"
    "github.com/kackerx/crontab/internal/worker/store/etcd"
    "github.com/kackerx/crontab/pkg/common"
    "github.com/kackerx/crontab/pkg/protocol"
    "github.com/kackerx/crontab/pkg/util"
    "strings"
)

type JobMgr struct {
    *etcd.Datastore
    Scheduler *scheduler.Scheduler
}

func NewJobMgr() *JobMgr {
    return &JobMgr{
        Datastore: nil,
    }
}

func (jm *JobMgr) WatchJobs() error {
    // 1, 获取所有/cron/jobs, 并且获取当前集群的revision
    getResp, err := jm.Kv.Get(context.Background(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
    if err != nil {
        return err
    }

    for _, kvPair := range getResp.Kvs {
        job, err := util.UnPackJob(kvPair.Value)
        if err != nil {
            return err
        }

        // todo: job同步给scheduler(调度协程)
        jobEvent := protocol.NewJobEvent(common.JOB_EVENT_SAVE, job)
        jm.Scheduler.PushJobEvent(jobEvent)
    }

    // 2, 从该revision监听当前变化
    go func() {
        wathchStartRevision := getResp.Header.Revision + 1
        watchCh := jm.Watcher.Watch(context.Background(), common.JOB_SAVE_DIR, clientv3.WithRev(wathchStartRevision), clientv3.WithPrefix())
        for watchResp := range watchCh {
            for _, watchEvent := range watchResp.Events {
                var jobEvent *protocol.JobEvent
                switch watchEvent.Type {
                case mvccpb.PUT:
                    // todo: 反序列化job推送scheduler
                    job, err := util.UnPackJob(watchEvent.Kv.Value)
                    if err != nil {
                        continue
                    }

                    jobEvent = protocol.NewJobEvent(common.JOB_EVENT_SAVE, job)
                case mvccpb.DELETE:
                    // todo: 推送删除事件给scheduler
                    jobName := strings.TrimPrefix(string(watchEvent.Kv.Key), common.JOB_SAVE_DIR)
                    jobEvent = protocol.NewJobEvent(common.JOB_EVENT_DELETE, &protocol.Job{Name: jobName})
                }

                // todo: push事件给调度器
                jm.Scheduler.PushJobEvent(jobEvent)

            }
        }
    }()
    return nil
}

// 监听强杀任务
func (jm *JobMgr) WatchKiller() {
    go func() {
        // 获取监听chan
        watchCh := jm.Watcher.Watch(context.TODO(), common.JOB_KILL_DIR, clientv3.WithPrefix())
        for watchResp := range watchCh {
            // 监听监听事件
            for _, watchEvent := range watchResp.Events {
                switch watchEvent.Type {
                case mvccpb.PUT: // 新增kill事件
                    jobName := strings.TrimPrefix(string(watchEvent.Kv.Key), common.JOB_KILL_DIR)
                    jobEvent := protocol.NewJobEvent(common.JOB_EVENT_KILL, &protocol.Job{Name: jobName})
                    scheduler.G_scheduler.PushJobEvent(jobEvent) // 生成job事件, 推送给调度器
                    fmt.Println(`kill job name: `, jobName)
                }
            }
        }
    }()
}
