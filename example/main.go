package main

import (
	"log"
	"time"

	"github.com/kuxuee/logger"
)

func main() {

	rotatingHandler := logger.NewRotatingHandler(".", "test", 1*1024*1024)

	// logger set handlers: console, rotating
	logger.SetHandlers(logger.Console, rotatingHandler)

	defer logger.Close()

	// logger set flags
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// logger set log level
	logger.SetLevel(logger.INFO)

	for i := 0; i < 100; i++ {
		logger.Debug("something1", "debug")
		logger.Info("something:", i)
		logger.Warn("something")
		logger.Error("something")
		time.Sleep(1 * time.Second)
	}

}
