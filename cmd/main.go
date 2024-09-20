package main

import (
	"os"
	"os/signal"
	"syscall"
	"test/internal/service"
	"test/internal/utils"
)

func main() {
	utils.LogMessage("Application run")

	app := service.NewApp()

	if err := app.Run(); err != nil {
		utils.LogMessage("Failed to run application: %v", err)
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	app.Wait()

	app.Cleanup()
}