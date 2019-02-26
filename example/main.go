package main

import (
	"log"
	"time"

	"github.com/kuxuee/logger"
)

func main() {
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
