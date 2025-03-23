package assignment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yashkumarverma/schedulerx/src/command"
	"github.com/yashkumarverma/schedulerx/src/utils"
	"github.com/yashkumarverma/schedulerx/src/utils/cache"
)

// Manager handles job assignments to pods
type Manager struct {
	redisClient *cache.Client
	logger      *utils.StandardLogger
	config      *utils.Config
}

// NewManager creates a new assignment manager
func NewManager(redisClient *cache.Client, logger *utils.StandardLogger, config *utils.Config) *Manager {
	return &Manager{
		redisClient: redisClient,
		logger:      logger,
		config:      config,
	}
}

// AssignJobs assigns unassigned jobs to available pods in a round-robin fashion
func (m *Manager) AssignJobs(ctx context.Context, pods []string) error {
	if len(pods) == 0 {
		return fmt.Errorf("no pods available for job assignment")
	}

	// Get the number of jobs to assign from config
	jobCount := m.config.NextJobCount
	if jobCount <= 0 {
		jobCount = 3 // Default value if not set
	}

	// Get unassigned jobs from Redis sorted set
	jobs, err := m.redisClient.GetClient().ZRange(ctx, command.JobsSortedSetKey, 0, int64(jobCount)-1).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %w", err)
	}

	// Round-robin assignment
	for i, jobID := range jobs {
		podIndex := i % len(pods)
		podID := pods[podIndex]

		// Get job details
		jobKey := fmt.Sprintf(command.JobDetailsKey, jobID)
		jobData, err := m.redisClient.GetClient().Get(ctx, jobKey).Bytes()
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
		if err := job.StoreInRedis(ctx, m.redisClient.GetClient()); err != nil {
			m.logger.Error("Failed to update job assignment", "job_id", job.ID, "pod_id", podID, "error", err)
			continue
		}

		m.logger.Info("Assigned job to pod", "job_id", job.ID, "pod_id", podID)
	}

	return nil
}

// UnassignJobsFromPod marks all jobs assigned to a specific pod as unassigned
func (m *Manager) UnassignJobsFromPod(ctx context.Context, podID string) error {
	// Get all jobs from Redis
	jobs, err := m.redisClient.GetClient().ZRange(ctx, command.JobsSortedSetKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %w", err)
	}

	for _, jobID := range jobs {
		jobKey := fmt.Sprintf(command.JobDetailsKey, jobID)
		jobData, err := m.redisClient.GetClient().Get(ctx, jobKey).Bytes()
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
			if err := job.StoreInRedis(ctx, m.redisClient.GetClient()); err != nil {
				m.logger.Error("Failed to unassign job", "job_id", job.ID, "pod_id", podID, "error", err)
				continue
			}

			m.logger.Info("Unassigned job from pod", "job_id", job.ID, "pod_id", podID)
		}
	}

	return nil
}
