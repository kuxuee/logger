package main

import (
	"log"

	"github.com/kuxuee/logger"
)

func main() {
	rotatingHandler := logger.NewRotatingHandler("./", "test.log", 4, 4*1024*1024)

	// logger set handlers: console, rotating
	logger.SetHandlers(logger.Console, rotatingHandler)

	defer logger.Close()

	// logger set flags
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// logger set log level
	logger.SetLevel(logger.INFO)

	logger.Debug("something1", "debug")
	logger.Info("something")
	logger.Warn("something")
	logger.Error("something")
}
