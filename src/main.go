package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yashkumarverma/schedulerx/src/leader"
	"github.com/yashkumarverma/schedulerx/src/scheduler"
	"github.com/yashkumarverma/schedulerx/src/utils"
	"github.com/yashkumarverma/schedulerx/src/utils/cache"
)

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	logger := utils.NewLogger()

	// Load configuration
	config := utils.GetConfig(ctx)

	// Create Redis client
	redisClient, err := cache.NewClient(ctx, config)
	if err != nil {
		logger.Fatal("Failed to create Redis client", err)
	}

	// Create command registry
	cmdRegistry := scheduler.NewCommandRegistry()

	// Display supported commands
	fmt.Println("\nSupported Linux Commands:")
	fmt.Println("------------------------")
	for cmd, desc := range cmdRegistry.GetCommandDescriptions() {
		fmt.Printf("%-15s - %s\n", cmd, desc)
	}
	fmt.Println("------------------------\n")

	// Initialize pod manager
	podManager := leader.NewPodManager(redisClient, logger, config)
	if err := podManager.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize pod manager", err)
	}

	logger.Info("Pod manager initialized successfully", "pod_id", podManager.GetPodID())

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down gracefully...")
}
