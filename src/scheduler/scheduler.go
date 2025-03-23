package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
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

// ScheduleJobs schedules the next batch of jobs
func (s *Scheduler) ScheduleJobs(ctx context.Context) error {
	if !leader.IsLeader() {
		fmt.Println("Not the leader, skipping scheduling")
		return nil
	}

	s.logger.Info("Scheduling jobs for all registered commands")

	// Get current time and end of scheduling window
	now := time.Now()
	endTime := now.Add(SchedulingWindow)

	// Get all commands from registry
	cmdRegistry := command.NewCommandRegistry()
	commands := cmdRegistry.GetCommands()

	// For each command, find execution times in the window
	for cmdID, cmd := range commands {
		scheduleStr, params, err := cmd.Schedule()
		if err != nil {
			s.logger.Error("Failed to get schedule for command", "command", cmdID, "error", err)
			continue
		}

		// Parse cron expression
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(scheduleStr)
		if err != nil {
			s.logger.Error("Failed to parse cron expression", "command", cmdID, "error", err)
			continue
		}

		// Get next execution times until end of window
		next := schedule.Next(now)
		for next.Before(endTime) {
			// Create job
			job := command.NewJob(cmdID, params, next)

			// Store job in Redis
			if err := job.StoreInRedis(ctx, s.redisClient.GetClient()); err != nil {
				s.logger.Error("Failed to store job", "job_id", job.ID, "error", err)
				continue
			}

			next = schedule.Next(next)
		}
	}

	// Fetch and print top 10 upcoming jobs
	fmt.Println("Fetching upcoming jobs")
	jobs, err := s.redisClient.GetClient().ZRange(ctx, command.JobsSortedSetKey, 0, 9).Result()
	fmt.Println("Jobs fetched", jobs)
	if err != nil {
		s.logger.Error("Failed to fetch upcoming jobs", "error", err)
		return nil
	}

	fmt.Println("\nTop 10 Upcoming Jobs:")
	fmt.Println("--------------------")
	for _, jobID := range jobs {
		jobKey := fmt.Sprintf(command.JobDetailsKey, jobID)
		jobData, err := s.redisClient.GetClient().Get(ctx, jobKey).Bytes()
		if err != nil {
			continue
		}

		var job command.Job
		if err := json.Unmarshal(jobData, &job); err != nil {
			continue
		}

		fmt.Printf("Job ID: %s\nCommand: %s\nScheduled: %s\nStatus: %s\nAssignedTo: %s\n--------------------\n",
			job.ID, job.CommandID, job.ScheduledAt.Format(time.RFC3339), job.Status, job.AssignedTo)
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

// AssignJobs assigns unassigned jobs to available pods in a round-robin fashion
func (s *Scheduler) AssignJobs(ctx context.Context, pods []string) error {
	if len(pods) == 0 {
		return fmt.Errorf("no pods available for job assignment")
	}

	// Get the number of jobs to assign from config
	jobCount := s.config.NextJobCount
	if jobCount <= 0 {
		jobCount = 3 // Default value if not set
	}

	// Get unassigned jobs from Redis sorted set
	jobs, err := s.redisClient.GetClient().ZRange(ctx, command.JobsSortedSetKey, 0, int64(jobCount)-1).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %w", err)
	}

	// Round-robin assignment
	for i, jobID := range jobs {
		podIndex := i % len(pods)
		podID := pods[podIndex]

		// Get job details
		jobKey := fmt.Sprintf(command.JobDetailsKey, jobID)
		jobData, err := s.redisClient.GetClient().Get(ctx, jobKey).Bytes()
		if err != nil {
			continue
		}

		var job command.Job
		if err := json.Unmarshal(jobData, &job); err != nil {
			continue
		}

		// Skip if job is already assigned or running
		if job.AssignedTo != "" || job.Status == command.Running {
			continue
		}

		// Update job with pod assignment
		job.AssignedTo = podID
		job.Status = command.Assigned

		// Store updated job in Redis
		if err := job.StoreInRedis(ctx, s.redisClient.GetClient()); err != nil {
			s.logger.Error("Failed to update job assignment", "job_id", job.ID, "pod_id", podID, "error", err)
			continue
		}

		s.logger.Info("Assigned job to pod", "job_id", job.ID, "pod_id", podID)
	}

	return nil
}

// UnassignJobsFromPod marks all jobs assigned to a specific pod as unassigned
func (s *Scheduler) UnassignJobsFromPod(ctx context.Context, podID string) error {
	// Get all jobs from Redis
	jobs, err := s.redisClient.GetClient().ZRange(ctx, command.JobsSortedSetKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %w", err)
	}

	for _, jobID := range jobs {
		jobKey := fmt.Sprintf(command.JobDetailsKey, jobID)
		jobData, err := s.redisClient.GetClient().Get(ctx, jobKey).Bytes()
		if err != nil {
			continue
		}

		var job command.Job
		if err := json.Unmarshal(jobData, &job); err != nil {
			continue
		}

		// Only unassign jobs that are assigned to this pod and not running
		if job.AssignedTo == podID && job.Status != command.Running {
			job.AssignedTo = ""
			job.Status = command.Scheduled

			// Store updated job in Redis
			if err := job.StoreInRedis(ctx, s.redisClient.GetClient()); err != nil {
				s.logger.Error("Failed to unassign job", "job_id", job.ID, "pod_id", podID, "error", err)
				continue
			}

			s.logger.Info("Unassigned job from pod", "job_id", job.ID, "pod_id", podID)
		}
	}

	return nil
}
