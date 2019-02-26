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

#配置文件logs.config
* name:单个logger配置项名字，由函数logger.NewLogger传入该名字作为参数来读取配置
* levle:日志级别0-debug 1-info 2-warm 3-error
* data:同一日志配置多个输出端
** handle:输出端console-控制台 file-普通文件 rotating-切片文件
** dir:切片文件目录
** filename:切片文件名,无须后缀名
** maxnum:最大支持文件数,达到设置值后向前覆盖文件
** maxsize:单个文件大小,达到大小后切片写新日志
```logs.config
{
	"logs" : [{
		"name":"default", 
		"level":0,
		"data":[
			{"handle":"console"},
			{"handle":"rotating", "dir":"./log", "filename":"default", "maxnum":0, "maxsize":"1MB"}
			]
	}]
}
```

```go
package main

import (
	"log"
	"time"

	"github.com/kuxuee/logger"
)

func main() {
	//default对应配置文件中name字段的值
	err := logger.NewLogger("default")
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Close()

	for i := 0; i < 10; i++ {
		logger.Debug("something1", "debug")
		logger.Info("something:", i)
		logger.Warn("something")
		logger.Error("something")
		time.Sleep(1 * time.Second)
	}
}

```
