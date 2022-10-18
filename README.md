# logger
logger是一个GO语言编写的简单日志库

# 特点
* 支持日志级别: DEBUG, INFO, WARN, ERROR, PANIC, FATAL
* 支持同时输出日志到控制台及文件:ConsoleHander, FileHandler, RotatingHandler
* 支持文件切片
* 基于golang基础包-log包开发

# 安装
```go
go get github.com/kuxuee/logger
```

# 配置文件logs.config
* name:单个logger配置项名字，由函数logger.NewLogger传入该名字作为参数来读取配置
* data:同一日志配置多个输出端
	* handle:输出端console-控制台 file-普通文件 rotating-切片文件
	* dir:切片文件目录
	* filename:切片文件名,无须后缀名
	* levle:日志级别0-debug 1-info 2-warn 3-error 4-panic 5-fatal
	* maxnum:最大支持文件数,达到设置值后向前覆盖文件,0为无限个
	* maxsize:单个文件大小,达到大小后切片写新日志
```logs.config
{
	"logs" : [{
	"name":"default", 
	"data":[
		{"handle":"console", "level":3},
		{"handle":"rotating", "dir":"./log", "filename":"default", "level":0, "maxnum":0, "maxsize":"1MB"}
		]
	}]
}
```

# 代码
```go
package main

import (
	"log"
	"time"

	"github.com/kuxuee/logger"
)

func main() {
	err := logger.NewLogger("./logs.config", "default")
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Close()

	for i := 0; i < 10; i++ {
		logger.Debug("something1", "debug")
		logger.Info("something:", i)
		logger.Warn("something")
		logger.Error("something")
		logger.Infof("This is info:%s-%d", "go", 11)
		time.Sleep(1 * time.Second)
		if 5 == i {
			logger.Fatal("fatal")
		}
	}
}

```
