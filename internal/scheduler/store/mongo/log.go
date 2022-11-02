package mongo

import (
    "context"
    "github.com/kackerx/crontab/internal/pkg/logsink"
    "github.com/kackerx/crontab/pkg/protocol"
)

type logRepo struct {
    data *Data
}

func (log *logRepo) InsertMany(batch *protocol.LogBatch) {
    log.data.collection.InsertMany(context.TODO(), batch.Logs)
}

func NewLogRepo(data *Data) logsink.LogRepo {
    return &logRepo{data: data}
}
