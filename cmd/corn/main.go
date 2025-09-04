package main

import (
	"log"
	"os"
	"time"

	"github.com/aziyan99/corn/internal/scheduler"
)

func main() {
	log.Println("INFO: Starting corn daemon...")

	configPath := "corntab"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	log.Printf("INFO: Loading configuration from '%s'", configPath)
	config, err := scheduler.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	log.Printf("INFO: Configuration loaded. Found %d jobs.", len(config.Jobs))

	go func() {
		now := time.Now()
		nextTick := now.Truncate(time.Minute).Add(time.Minute)
		time.Sleep(nextTick.Sub(now))

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			checkJobs(config.Jobs)
		}
	}()

	select {}
}

func checkJobs(jobs []scheduler.Job) {
	currentTime := time.Now()
	log.Printf("INFO: Corn daemon tick at %s", currentTime.Format(time.RFC1123))
	for _, job := range jobs {
		jobToRun := job
		if jobToRun.ShouldRun(currentTime) {
			go jobToRun.Run()
		}
	}
}
