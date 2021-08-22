package main

import (
	"crontab/master/configs"
	"crontab/master/internal"
	"crontab/master/pkg"
	"flag"
	"log"
	"runtime"
)

var configFile string

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initArgs() {
	flag.StringVar(&configFile, "config", "./configs/config.json", "配置文件路径")
	flag.Parse()
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

	// 初始化任务处理器
	pkg.G_jobMgr, err = pkg.InitJobMgr(pkg.G_config.EtcdConfig)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	// 运行服务器
	err := internal.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
