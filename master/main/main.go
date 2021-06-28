package main

import (
	"flag"
	"fmt"
	"github.com/kackerx/crontab/master"
	"runtime"
	"time"
)

var (
	confFile string // 配置文件路径
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initArgs() {
	flag.StringVar(&confFile, "config", "./master.json", "配置文件路径")
	flag.Parse()
}

func main() {
	var (
		err error
	)

	// 初始化线程
	initEnv()

	// 加载配置
	initArgs()
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}

	// 启动任务管理器
	if err = master.InitJobMar(); err != nil {
		goto ERR
	}

	// 启动HTTP服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Println("main")
	return

ERR:
	fmt.Println(err)
}