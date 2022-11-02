package protocol

import (
    "context"
    "fmt"
    "github.com/gorhill/cronexpr"
    "time"
)

// 定时任务
type Job struct {
    Name     string `json:"name"`     // 任务名
    Command  string `json:"command"`  // shell命令
    CronExpr string `json:"cronExpr"` // cron表达式
}

// HTTP接口应答
type Response struct {
    Errno int         `json:"errno"`
    Msg   string      `json:"msg"`
    Data  interface{} `json:"data"`
}

//变化事件
type JobEvent struct {
    EventType int // save, delete
    Job       *Job
}

func NewJobEvent(eventType int, job *Job) *JobEvent {
    return &JobEvent{
        EventType: eventType,
        Job:       job,
    }
}

type JobSchedulePlan struct {
    Job  *Job
    Expr *cronexpr.Expression
    Next time.Time
}

func NewJobSchedulePlan(job *Job) (*JobSchedulePlan, error) {
    cron := job.CronExpr
    //cron = "*/3 * * * *"
    expr, err := cronexpr.Parse(cron)
    if err != nil {
        fmt.Println("解析错误: ", err)
        return nil, err
    }

    return &JobSchedulePlan{
        Job:  job,
        Expr: expr,
        Next: expr.Next(time.Now()),
    }, nil
}

type JobExecuteInfo struct {
    Job        *Job
    PlanTime   time.Time          // 理论时间
    RealTime   time.Time          // 真实时间
    CancelCtx  context.Context    // 用于取消任务
    CancelFunc context.CancelFunc // 用于取消任务的函数
}

func NewJobExecuteInfo(plan *JobSchedulePlan) *JobExecuteInfo {
    ctx, cancel := context.WithCancel(context.TODO())
    return &JobExecuteInfo{Job: plan.Job, PlanTime: plan.Next, RealTime: time.Now(), CancelCtx: ctx, CancelFunc: cancel}
}

type JobExecuteResult struct {
    ExecuteInfo *JobExecuteInfo
    OutPut      []byte
    Err         error
    StartTime   time.Time
    EndTime     time.Time
}

// 任务执行日志
type JobLog struct {
    JobName      string `bson:"jobName"`
    Command      string `bson:"command"`
    Err          string `bson:"err"`
    Output       string `bson:"output"`
    PlanTime     int64  `bson:"planTime"`
    ScheduleTime int64  `bson:"scheduleTime"`
    StartTime    int64  `bson:"startTime"`
    EndTime      int64  `bson:"endTime"`
}

type LogBatch struct {
    Logs []interface{}
}

//func NewJobExecuteResult(executeInfo JobExecuteInfo, outPut []byte, err error, startTime time.Time, endTime time.Time) *JobExecuteResult {
//	return &JobExecuteResult{ExecuteInfo: executeInfo, OutPut: outPut, Err: err, StartTime: startTime, EndTime: endTime}
//}
