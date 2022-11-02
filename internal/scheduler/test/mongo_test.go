package test

import (
    "fmt"
    "github.com/kackerx/crontab/internal/scheduler/config"
    "github.com/kackerx/crontab/internal/scheduler/store/mongo"
    "testing"
)

func TestMongo(t *testing.T) {
    cfg, err := config.NewConfig()
    if err != nil {
        fmt.Printf("%+v\n", err)
        return
    }

    db, err := mongo.NewData(cfg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", db)
}

func TestFoo(t *testing.T) {
    a := []string{"k"}
    a = a[:0]

    fmt.Println(a)
}
