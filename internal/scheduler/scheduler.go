package scheduler

import (
    "fmt"
    "github.com/kackerx/crontab/internal/pkg/logsink"
    "github.com/kackerx/crontab/pkg/common"
    "github.com/kackerx/crontab/pkg/protocol"
    "time"
)

var (
    G_scheduler *Scheduler
)

type Scheduler struct {
    jobEventChan       chan *protocol.JobEvent `json:"test"`
    jobExecuteResultCh chan *protocol.JobExecuteResult
    jobPlanTable       map[string]*protocol.JobSchedulePlan
    jobExecutingTable  map[string]*protocol.JobExecuteInfo
}

func NewScheduler() *Scheduler {
    return &Scheduler{
        jobEventChan:       make(chan *protocol.JobEvent, 1000),
        jobPlanTable:       make(map[string]*protocol.JobSchedulePlan),
        jobExecutingTable:  make(map[string]*protocol.JobExecuteInfo),
        jobExecuteResultCh: make(chan *protocol.JobExecuteResult, 1000),
    }
}

func (scheduler *Scheduler) PushJobEvent(event *protocol.JobEvent) {
    scheduler.jobEventChan <- event
}

func (scheduler *Scheduler) handleJobEvent(jobEvent *protocol.JobEvent) {
    switch jobEvent.EventType {
    case common.JOB_EVENT_SAVE:
        JobSchedulePlan, err := protocol.NewJobSchedulePlan(jobEvent.Job)
        if err != nil {
            return
        }

        scheduler.jobPlanTable[jobEvent.Job.Name] = JobSchedulePlan
    case common.JOB_EVENT_DELETE:
        fmt.Println("删除任务", jobEvent.Job.Name)
        if _, ok := scheduler.jobPlanTable[jobEvent.Job.Name]; ok {
            delete(scheduler.jobPlanTable, jobEvent.Job.Name)
        }
    case common.JOB_EVENT_KILL:
        fmt.Printf("强杀任务: %s, %v\n", jobEvent.Job.Name, scheduler.jobExecutingTable)
        if jobExecuteinfo, ok := scheduler.jobExecutingTable[jobEvent.Job.Name]; ok {
            fmt.Printf("任务%s正在执行中, 执行强杀!\n", jobEvent.Job.Name)
            jobExecuteinfo.CancelFunc()
        }

    }

}

func (scheduler *Scheduler) ScheduleLoop() {
    scheduleAfter := scheduler.TrySchedule()
    scheduleTimer := time.NewTimer(scheduleAfter)

    for {
        select {
        case jobEvent := <-scheduler.jobEventChan:
            // 监听任务变化, 对任务列表做增删改查
            scheduler.handleJobEvent(jobEvent)
        case result := <-scheduler.jobExecuteResultCh:
            scheduler.handleJobResult(result)
        case <-scheduleTimer.C:
        }

        scheduleAfter = scheduler.TrySchedule()
        scheduleTimer.Reset(scheduleAfter)
    }

}

func (scheduler *Scheduler) TrySchedule() time.Duration {
    var nearTime *time.Time

    if len(scheduler.jobPlanTable) == 0 {
        return time.Second * 1
    }

    now := time.Now()
    for _, jobPlan := range scheduler.jobPlanTable {
        if jobPlan.Next.Before(now) || jobPlan.Next.Equal(now) {
            // TODO: 尝试执行任务
            scheduler.TryStartJob(jobPlan)
            jobPlan.Next = jobPlan.Expr.Next(now)
            //fmt.Printf("执行任务%s: %s\n", jobPlan.Job.Name, time.Now().String())
        }

        if nearTime == nil || jobPlan.Next.Before(*nearTime) {
            nearTime = &jobPlan.Next
        }
    }

    return (*nearTime).Sub(now)
}

func (scheduler *Scheduler) TryStartJob(plan *protocol.JobSchedulePlan) {

    var (
        jobExecuteInfo *protocol.JobExecuteInfo
    )

    //fmt.Printf("before ExecutingTable: %s\n", scheduler.jobExecutingTable)

    if _, ok := scheduler.jobExecutingTable[plan.Job.Name]; ok {
        // 任务已经在运行中, 去重
        fmt.Printf("跳过执行中的任务: %s\n", plan.Job.Name)
        return
    }

    jobExecuteInfo = protocol.NewJobExecuteInfo(plan)
    scheduler.jobExecutingTable[plan.Job.Name] = jobExecuteInfo

    //fmt.Printf("after ExecutingTable: %s\n", scheduler.jobExecutingTable)

    // @ 执行任务
    go G_executor.ExecuteJob(jobExecuteInfo)
    //fmt.Printf("开始执行: %s: %s: %s\n", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
}

func (scheduler *Scheduler) PushJobResult(result *protocol.JobExecuteResult) {
    scheduler.jobExecuteResultCh <- result
}

func (scheduler *Scheduler) handleJobResult(result *protocol.JobExecuteResult) {
    delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

    // 非抢锁失败, 记录日志
    if result.Err != common.ERR_LOCKED {
        jobLog := &protocol.JobLog{
            JobName:      result.ExecuteInfo.Job.Name,
            Command:      result.ExecuteInfo.Job.Command,
            Output:       string(result.OutPut),
            PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
            ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
            StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
            EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
        }

        if result.Err != nil {
            jobLog.Err = result.Err.Error()
        } else {
            jobLog.Err = ""
        }

        logsink.Logsink.Append(jobLog)
    }

    //fmt.Printf("任务执行完成: %s: %s\n", result.ExecuteInfo.Job.Name, result.OutPut)

}
