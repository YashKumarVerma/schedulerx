package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	Scheduled JobStatus = "scheduled"
	Running   JobStatus = "running"
	Failed    JobStatus = "failed"
	Success   JobStatus = "success"
)

// Redis keys
const (
	JobsSortedSetKey = "scheduler:jobs"
	JobDetailsKey    = "scheduler:job:%s" // Format string for job details
)

// Job represents a scheduled command execution
type Job struct {
	ID          string     // Unique Job ID
	CommandID   string     // Related Command
	Params      []string   // Command parameters
	Status      JobStatus  // Current status of the job
	ScheduledAt time.Time  // When the job is scheduled to run
	StartedAt   *time.Time // When the job actually started
	FinishedAt  *time.Time // When the job finished
	Error       string     // Error message if job failed
}

// NewJob creates a new job with a unique ID based on command ID and scheduled time
func NewJob(commandID string, params []string, scheduledAt time.Time) *Job {
	// Create a unique ID by combining command ID and scheduled time
	// Format: commandID_timestamp
	jobID := fmt.Sprintf("%s_%d", commandID, scheduledAt.Unix())

	return &Job{
		ID:          jobID,
		CommandID:   commandID,
		Params:      params,
		Status:      Scheduled,
		ScheduledAt: scheduledAt,
	}
}

// StoreInRedis stores the job in Redis using a sorted set for scheduling and a hash for job details
func (j *Job) StoreInRedis(ctx context.Context, client *redis.Client) error {
	// Store job details in a hash
	jobKey := fmt.Sprintf(JobDetailsKey, j.ID)
	jobData, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Store in sorted set with scheduled time as score
	pipe := client.Pipeline()
	pipe.Set(ctx, jobKey, jobData, 24*time.Hour) // Store for 24 hours
	pipe.ZAdd(ctx, JobsSortedSetKey, redis.Z{
		Score:  float64(j.ScheduledAt.Unix()),
		Member: j.ID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store job in Redis: %w", err)
	}

	return nil
}

// UpdateInRedis updates the job status and details in Redis
func (j *Job) UpdateInRedis(ctx context.Context, client *redis.Client) error {
	jobKey := fmt.Sprintf(JobDetailsKey, j.ID)
	jobData, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	pipe := client.Pipeline()

	// Update job details
	pipe.Set(ctx, jobKey, jobData, 24*time.Hour)

	// If job is completed (success or failed), remove from sorted set
	if j.Status == Success || j.Status == Failed {
		pipe.ZRem(ctx, JobsSortedSetKey, j.ID)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update job in Redis: %w", err)
	}

	return nil
}

// Start marks the job as running and sets the start time
func (j *Job) Start() {
	now := time.Now()
	j.StartedAt = &now
	j.Status = Running
}

// Complete marks the job as successful and sets the finish time
func (j *Job) Complete() {
	now := time.Now()
	j.FinishedAt = &now
	j.Status = Success
}

// Fail marks the job as failed, sets the finish time and error message
func (j *Job) Fail(err error) {
	now := time.Now()
	j.FinishedAt = &now
	j.Status = Failed
	if err != nil {
		j.Error = err.Error()
	}
}

// IsOverdue checks if the job is overdue based on its scheduled time
func (j *Job) IsOverdue() bool {
	return time.Now().After(j.ScheduledAt)
}

// Duration returns the duration of the job execution if it has finished
func (j *Job) Duration() *time.Duration {
	if j.StartedAt == nil || j.FinishedAt == nil {
		return nil
	}
	duration := j.FinishedAt.Sub(*j.StartedAt)
	return &duration
}

// String returns a string representation of the job
func (j *Job) String() string {
	return fmt.Sprintf("Job[%s] - Command: %s, Status: %s, Scheduled: %s",
		j.ID, j.CommandID, j.Status, j.ScheduledAt.Format(time.RFC3339))
}
