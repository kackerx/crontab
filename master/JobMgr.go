package master

import (
    "context"
    "encoding/json"
    "github.com/coreos/etcd/clientv3"
    "github.com/kackerx/crontab/common"
    "time"
)

var (
    G_jobMgr *JobMgr
)

type JobMgr struct {
    client *clientv3.Client
    kv     clientv3.KV
    lease  clientv3.Lease
}

func InitJobMar() (err error) {
    var (
        config clientv3.Config
        client *clientv3.Client
        kv     clientv3.KV
        lease  clientv3.Lease
    )
    // 初始化配置
    config = clientv3.Config{
        Endpoints:   G_config.EtcdEndpoints,                                     // 集群地址
        DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, // 连接超时
    }
    
    // 建立连接
    if client, err = clientv3.New(config); err != nil {
        return
    }
    
    // kv和lease
    kv = clientv3.NewKV(client)
    lease = clientv3.NewLease(client)
    
    G_jobMgr = &JobMgr{
        client: client,
        kv:     kv,
        lease:  lease,
    }
    return
}

func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
    var (
        jobKey   string
        jobValue []byte
        putResp  *clientv3.PutResponse
        //oldJobObj common.Job
    )
    // etcd的保存key
    jobKey = "/cron/jobs/" + job.Name
    // 任务信息json
    if jobValue, err = json.Marshal(job); err != nil {
        return
    }
    // 保存到etcd
    if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
        return
    }
    // 如果是更新, 返回旧值
    if putResp.PrevKv != nil {
        // 旧值反序列化
        if err = json.Unmarshal(putResp.PrevKv.Value, &oldJob); err != nil {
            return
        }
        //oldJob = &oldJobObj
        return
    }
    return
}

func (jobMgr *JobMgr) DeleteJob(job *common.Job) (oldJob *common.Job, err error) {
    var (
        jobKey     string
        deleteResp *clientv3.DeleteResponse
    )
    jobKey = "/cron/jobs/" + job.Name
    if deleteResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
        return
    }
    // 返回被删除的任务
    if len(deleteResp.PrevKvs) != 0 {
        if err = json.Unmarshal(deleteResp.PrevKvs[0].Value, &oldJob); err != nil {
            return
        }
    }
    return
}
