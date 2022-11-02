package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/kackerx/crontab/internal/pkg/logsink"
	"github.com/kackerx/crontab/internal/scheduler"
	config2 "github.com/kackerx/crontab/internal/scheduler/config"
	"github.com/kackerx/crontab/internal/scheduler/store/mongo"
	"github.com/kackerx/crontab/internal/worker/config"

	"github.com/kackerx/crontab/internal/worker/options"
	"github.com/kackerx/crontab/internal/worker/service"
	"github.com/kackerx/crontab/internal/worker/store/etcd"
	"time"
)

func main() {
	// 1, 初始化配置, 创建etcd依赖
	options := options.NewOptions()

	cfg, _ := config.NewConfig(options)

	client, err := clientv3.New(clientv3.Config{Endpoints: cfg.EtcdOptions.Endpoints, DialTimeout: time.Duration(cfg.EtcdOptions.Timeout) * time.Millisecond})
	if err != nil {
		panic(err)
	}

	// 初始化etcd连接
	data := etcd.NewDatastore(client, client.KV, client.Lease, client.Watcher)
	scheduler.G_scheduler = scheduler.NewScheduler()
	jobMgr := &service.JobMgr{
		Datastore: data,
		Scheduler: scheduler.G_scheduler,
	}

	// 初始化分布式锁
	scheduler.GLock = scheduler.NewLock("", client)

	// 初始化执行器
	scheduler.G_executor = scheduler.NewExecutor()

	// 监听任务变化
	if err := jobMgr.WatchJobs(); err != nil {
		panic(err)
	}
	fmt.Println("监听开始")

	// 监听任务变化
	jobMgr.WatchKiller()
	fmt.Println("监听killer")

	// 日志记录器
	config, err := config2.NewConfig()
	if err != nil {
		panic(err)
	}

	mongoData, err := mongo.NewData(config)
	if err != nil {
		panic(err)
	}

	logRepo := mongo.NewLogRepo(mongoData)
	logsink.InitLogSink(logsink.NewLogUsecase(logRepo))
	fmt.Println("启动日志")
	go logsink.Logsink.WriteLoop()

	// 调度器调度运行
	fmt.Println("调度开始")
	jobMgr.Scheduler.ScheduleLoop()

	time.Sleep(time.Second * 1000)

}
