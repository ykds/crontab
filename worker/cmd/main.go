package main

import (
	"crontab/worker/configs"
	"crontab/worker/internal"
	"crontab/worker/pkg"
	"flag"
	"fmt"
	"log"
	"runtime"
)

var configFile string

func initArgs() {
	flag.StringVar(&configFile, "config", "./configs/config.json", "配置文件路径")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func init() {
	initEnv()
	initArgs()

	var err error

	// 初始化配置
	pkg.G_config, err = configs.InitConfig(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	// 初始化任务调度器
	internal.InitScheduler()

	// 初始化任务执行器
	err = internal.InitExecutor()
	if err != nil {
		log.Fatalln(err)
	}

	// 初始化日志处理器
	err = pkg.InitLogSink()
	if err != nil {
		log.Fatalln(err)
	}

	// 初始化任务管理器
	pkg.G_jobMgr, err = pkg.InitJobMgr(pkg.G_config.EtcdConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func main() {
	// 开启调度器监听
	go pkg.G_scheduler.ScheduleLoop()
	// 开启日志事件监听
	go pkg.G_logSink.WriteLoop()

	// 开启任务事件监听
	err := pkg.G_jobMgr.WatchJobs()
	if err != nil {
		log.Fatalln(err)
	}
	// 开启任务杀死事件监听
	go pkg.G_jobMgr.WatchKillJob()

	select {}
}
