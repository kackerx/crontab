package mongo

import (
    "context"
    "fmt"
    "github.com/kackerx/crontab/internal/scheduler/config"
    "github.com/pkg/errors"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "time"
)

type Data struct {
    client     *mongo.Client
    collection *mongo.Collection
}

func NewData(cfg *config.Config) (*Data, error) {
    ctx := context.TODO()

    clientOptions := options.Client().
        ApplyURI(cfg.Uri).
        SetAuth(options.Credential{
            Username: "admin",
            Password: "123456",
        }).
        SetTimeout(time.Duration(cfg.Timeout) * time.Millisecond)

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, errors.Wrap(err, "连接mongodb失败")
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        panic(err)
    }
    fmt.Println("Connected to MongoDB!")

    collection := client.Database("cron").Collection("log")

    return &Data{
        client:     client,
        collection: collection,
    }, nil
}
