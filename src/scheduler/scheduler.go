package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/yashkumarverma/schedulerx/src/command"
	"github.com/yashkumarverma/schedulerx/src/leader"
	"github.com/yashkumarverma/schedulerx/src/utils"
	"github.com/yashkumarverma/schedulerx/src/utils/cache"
)

const (
	// SchedulingWindow is the time window for which we schedule jobs
	SchedulingWindow = 5 * time.Minute
)

// Scheduler handles job scheduling for the leader pod
type Scheduler struct {
	redisClient *cache.Client
	logger      *utils.StandardLogger
	config      *utils.Config
	commands    map[string]command.Command
}

// NewScheduler creates a new scheduler instance
func NewScheduler(redisClient *cache.Client, logger *utils.StandardLogger, config *utils.Config) *Scheduler {
	return &Scheduler{
		redisClient: redisClient,
		logger:      logger,
		config:      config,
		commands:    make(map[string]command.Command),
	}
}

// RegisterCommand adds a command to the scheduler
func (s *Scheduler) RegisterCommand(cmd command.Command) {
	s.commands[cmd.ID()] = cmd
}

// ScheduleJobs creates and stores jobs for all registered commands
func (s *Scheduler) ScheduleJobs(ctx context.Context) error {
	if !leader.IsLeader() {
		return nil
	}

	s.logger.Info("Scheduling jobs for all registered commands")

	now := time.Now()
	windowEnd := now.Add(SchedulingWindow)

	for _, cmd := range s.commands {
		// Get the next execution times for this command within the scheduling window
		nextTimes, err := s.getNextExecutionTimesInWindow(cmd, now, windowEnd)
		if err != nil {
			s.logger.Error("Failed to get next execution times for command", "command", cmd.ID(), "error", err)
			continue
		}

		// Create and store jobs for each execution time
		for _, nextTime := range nextTimes {
			job := command.NewJob(cmd.ID(), cmd.Parameters(), nextTime)
			if err := job.StoreInRedis(ctx, s.redisClient.GetClient()); err != nil {
				s.logger.Error("Failed to store job in Redis", "job", job.ID, "error", err)
				continue
			}
			s.logger.Info("Created and stored job", "job", job.ID, "command", cmd.ID(), "scheduled_at", nextTime)
		}
	}

	return nil
}

// getNextExecutionTimesInWindow calculates the next execution times for a command within a time window
func (s *Scheduler) getNextExecutionTimesInWindow(cmd command.Command, start, end time.Time) ([]time.Time, error) {
	nextTimes := make([]time.Time, 0)
	currentTime := start

	// Get the cron schedule for this command
	schedule, _, err := cmd.Schedule()
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule for command %s: %w", cmd.ID(), err)
	}

	// Parse the cron expression
	parser := NewParser()
	expr, err := parser.Parse(schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cron expression %s: %w", schedule, err)
	}

	// Find all execution times within the window
	for currentTime.Before(end) {
		nextTime := expr.Next(currentTime)
		if nextTime.IsZero() || nextTime.After(end) {
			break // No more future executions in the window
		}
		nextTimes = append(nextTimes, nextTime)
		currentTime = nextTime.Add(time.Second) // Move to next second to avoid duplicates
	}

	return nextTimes, nil
}

// Start begins the scheduler
func (s *Scheduler) Start() error {
	// TODO: Implement scheduler logic
	return nil
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() error {
	// TODO: Implement graceful shutdown
	return nil
}

// ScheduleTask schedules a task to run at a specific time
func (s *Scheduler) ScheduleTask(taskID string, executionTime time.Time) error {
	// TODO: Implement task scheduling
	return nil
}

// CancelTask cancels a scheduled task
func (s *Scheduler) CancelTask(taskID string) error {
	// TODO: Implement task cancellation
	return nil
}

// ListTasks returns all scheduled tasks
func (s *Scheduler) ListTasks() ([]string, error) {
	// TODO: Implement task listing
	return nil, nil
}
