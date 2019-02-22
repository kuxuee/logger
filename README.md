# logger
logger是一个GO语言编写的简单日志库

# 特点
* 支持日志级别: DEBUG, INFO, WARN, ERROR
* 支持同时输出日志到控制台及文件:ConsoleHander, FileHandler, RotatingHandler
* 基于golang基础包-log包开发

# 安装
```go
go get github.com/kuxuee/logger
```

#例子
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kuxuee/logger"
)

func main() {
	//日志输出目录
	//日志文件名，无须后缀
	//同一日志最大文件个数,达到个数后会自动往前覆盖,INFINITE为无限个
	//单个日志文件大小,达到日志大小后组件自动切割日志
	rotatingHandler, err := logger.NewRotatingHandler(".", "test", logger.INFINITE, 1*1024*1024)
	if err != nil {
		fmt.Printf("NewRotatingHandler error:%v", err)
		return
	}

	//设置同时输出到控制台及文件
	logger.SetHandlers(logger.Console, rotatingHandler)

	defer logger.Close()

	//设置日志标签
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)

	//设置日志级别
	logger.SetLevel(logger.INFO)

	for i := 0; i < 10; i++ {
		logger.Debug("something1", "debug")
		logger.Info("something:", i)
		logger.Warn("something")
		logger.Error("something")
		time.Sleep(1 * time.Second)
	}
}
```
