package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yashkumarverma/schedulerx/src/command"
	"github.com/yashkumarverma/schedulerx/src/leader"
	"github.com/yashkumarverma/schedulerx/src/scheduler"
	"github.com/yashkumarverma/schedulerx/src/utils"
	"github.com/yashkumarverma/schedulerx/src/utils/cache"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := utils.NewLogger()
	config := utils.GetConfig(ctx)

	redisClient, err := cache.NewClient(ctx, config)
	if err != nil {
		logger.Fatal("Failed to create Redis client", err)
	}

	cmdRegistry := command.NewCommandRegistry()

	// only hardcoded tasks supported now
	fmt.Println("\nSupported Commands:")
	for cmd, desc := range cmdRegistry.GetCommandDescriptions() {
		fmt.Printf("%-15s - %s\n", cmd, desc)
	}

	// Initialize pod manager
	podManager := leader.NewPodManager(redisClient, logger, config)
	if err := podManager.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize pod manager", err)
	}

	logger.Info("Pod manager initialized successfully", "pod_id", podManager.GetPodID())

	// Create scheduler instance
	scheduler := scheduler.NewScheduler(redisClient, logger, config)

	// Register all commands with the scheduler
	for cmdID, cmd := range cmdRegistry.GetCommands() {
		scheduler.RegisterCommand(cmd)
		logger.Info("Registered command with scheduler", "command", cmdID)
	}

	// Start job scheduling routine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := scheduler.ScheduleJobs(ctx); err != nil {
					logger.Error("Failed to schedule jobs", "error", err)
				}
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down gracefully...")
}
