package logsink

import (
    "fmt"
    "github.com/kackerx/crontab/pkg/protocol"
    "time"
)

var Logsink *LogSink

type LogSink struct {
    *LogUsecase
    logCh        chan *protocol.JobLog
    autoCommitCh chan *protocol.LogBatch
}

type LogRepo interface {
    InsertMany(*protocol.LogBatch)
}

type LogUsecase struct {
    repo LogRepo
}

func NewLogUsecase(repo LogRepo) *LogUsecase {
    return &LogUsecase{repo: repo}
}

func InitLogSink(usecase *LogUsecase) error {
    Logsink = &LogSink{usecase, make(chan *protocol.JobLog, 1000), make(chan *protocol.LogBatch, 1000)}
    return nil
}

func (ls *LogSink) saveJobs(batch *protocol.LogBatch) {
    ls.repo.InsertMany(batch)
}

func (ls *LogSink) WriteLoop() {
    var (
        logBatch    *protocol.LogBatch
        commitTimer *time.Timer
    )

    for {
        select {
        case log := <-ls.logCh:
            // 批量写入批次日志
            if logBatch == nil {
                logBatch = &protocol.LogBatch{}
                // 超时不足100自动提交
                commitTimer = time.AfterFunc(time.Second,
                    func(batch *protocol.LogBatch) func() {
                        return func() {
                            ls.autoCommitCh <- batch
                        }
                    }(logBatch),
                )
            }

            logBatch.Logs = append(logBatch.Logs, log)

            if len(logBatch.Logs) >= 100 { // 超过批量提交
                fmt.Println("批量提交日志")
                ls.saveJobs(logBatch)
                logBatch = nil
                commitTimer.Stop()
            }
        case timeoutBatch := <-ls.autoCommitCh:

            // 判断超时批次, 是否仍旧是当前旧批次
            if timeoutBatch != logBatch {
                continue
            }

            fmt.Println("写入超时ch")
            ls.saveJobs(timeoutBatch)
            logBatch = nil
        }
    }
}

func (ls *LogSink) Append(log *protocol.JobLog) {
    select {
    case ls.logCh <- log:
    default:
    }
}
