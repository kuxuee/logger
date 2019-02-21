package main

import (
	"log"

	"github.com/kuxuee/logger"
)

func main() {
	//.:日志输出目录
	//test:日志文件名，无须后缀
	//1*1024*1024:单个日志文件大小,达到日志大小后组件自动切割日志
	rotatingHandler := logger.NewRotatingHandler(".", "test", 1*1024*1024)

	//设置同时输出到控制台及文件
	logger.SetHandlers(logger.Console, rotatingHandler)

	defer logger.Close()

	//设置日志标签
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//设置日志级别
	logger.SetLevel(logger.INFO)

	logger.Debug("something1", "debug")
	logger.Info("something:")
	logger.Warn("something")
	logger.Error("something")
}
