package scheduler

import (
    "fmt"
    "github.com/kackerx/crontab/pkg/protocol"
    "math/rand"
    "os/exec"
    "time"
)

var (
    G_executor *Executor
)

type Executor struct {
}

func NewExecutor() *Executor {
    return &Executor{}
}

func (executor *Executor) ExecuteJob(info *protocol.JobExecuteInfo) {
    jobExecuteResult := protocol.JobExecuteResult{
        ExecuteInfo: info,
        OutPut:      make([]byte, 0),
        Err:         nil,
        StartTime:   time.Now(),
    }
    // 锁, 随机睡眠防止抢锁倾斜
    time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
    GLock.LockKey = info.Job.Name
    err := GLock.TryLock()
    defer GLock.UnLock()

    if err != nil {
        // 锁失败
        jobExecuteResult.Err = err
        jobExecuteResult.EndTime = time.Now()
    } else {
        cmd := exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)

        // 执行
        outPut, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println(err)
            //return
        }

        jobExecuteResult.EndTime = time.Now()
        jobExecuteResult.OutPut = outPut
        jobExecuteResult.Err = err

    }
    G_scheduler.PushJobResult(&jobExecuteResult)
}
