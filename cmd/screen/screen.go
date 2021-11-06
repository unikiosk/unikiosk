package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/unikiosk/unikiosk/pkg/config"
	"github.com/unikiosk/unikiosk/pkg/service"
	"github.com/unikiosk/unikiosk/pkg/util/logger"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	c, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.GetLoggerInstance("", logger.ParseLogLevel(c.LogLevel))

	kiosk, err := service.New(log, c)
	if err != nil {
		return err
	}

	err = kiosk.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
