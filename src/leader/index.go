package leader

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yashkumarverma/schedulerx/src/assignment"
	"github.com/yashkumarverma/schedulerx/src/utils"
	"github.com/yashkumarverma/schedulerx/src/utils/cache"
)

const (
	// Redis key for storing pod information
	podRegistryKey = "schedulerx:pods"

	// TTL for pod presence. if not heard for 5 minutes, assume pod to be dead
	podTTL = 5 * time.Second
)

type PodInfo struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	LastSeen  time.Time `json:"last_seen"`
	Status    string    `json:"status"`
	IsLeader  bool      `json:"is_leader""`
}

var (
	once     sync.Once
	instance *PodManager
)

// PodManager handles pod registration and presence updates
type PodManager struct {
	client     *cache.Client
	logger     *utils.StandardLogger
	config     *utils.Config
	info       *PodInfo
	assignment *assignment.Manager
}

// NewPodManager creates a new pod manager instance
func NewPodManager(client *cache.Client, logger *utils.StandardLogger, config *utils.Config) *PodManager {
	once.Do(func() {
		instance = &PodManager{
			client:     client,
			logger:     logger,
			config:     config,
			assignment: assignment.NewManager(client, logger, config),
		}
	})
	return instance
}

// Initialize sets up the pod with a unique ID and starts presence updates
func (pm *PodManager) Initialize(ctx context.Context) error {
	// get or create and ID for the current pod
	podID := pm.config.PodID
	if podID == "" {
		podID = uuid.New().String()
	}

	pm.info = &PodInfo{
		ID:        podID,
		StartTime: time.Now(),
		LastSeen:  time.Now(),
		Status:    "active",
		IsLeader:  false,
	}
	fmt.Printf("Current Pod ID: %s", pm.info.ID)

	if err := pm.registerPod(ctx); err != nil {
		return fmt.Errorf("failed to register pod: %w", err)
	}

	// start pod heartbeat
	go pm.startPresenceUpdates(ctx)

	pm.logger.Info("Pod manager initialized", "pod_id", podID)
	return nil
}

// registerPod registers the pod in Redis
func (pm *PodManager) registerPod(ctx context.Context) error {
	if pm.info == nil {
		return fmt.Errorf("pod info not initialized")
	}

	// Get existing pods
	pods, err := pm.getPods(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	// Add or update current pod
	pods[pm.info.ID] = PodInfo{
		ID:        pm.info.ID,
		StartTime: pm.info.StartTime,
		LastSeen:  pm.info.LastSeen,
		Status:    pm.info.Status,
	}

	// Store updated pods
	if err := pm.client.SetJSONWithExpiry(ctx, podRegistryKey, pods, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to store pods: %w", err)
	}

	return nil
}

// getPods retrieves all registered pods from Redis
func (pm *PodManager) getPods(ctx context.Context) (map[string]PodInfo, error) {
	var pods map[string]PodInfo
	if err := pm.client.GetJSON(ctx, podRegistryKey, &pods); err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}
	if pods == nil {
		return make(map[string]PodInfo), nil
	}
	return pods, nil
}

// cleanupDeadPods removes pods that haven't been seen for longer than podTTL
func (pm *PodManager) cleanupDeadPods(ctx context.Context, pods map[string]PodInfo) map[string]PodInfo {
	cleanedPods := make(map[string]PodInfo)
	now := time.Now()

	for id, info := range pods {
		if now.Sub(info.LastSeen) <= podTTL {
			cleanedPods[id] = info
		}
	}

	return cleanedPods
}

// startPresenceUpdates begins the routine to update pod presence
func (pm *PodManager) startPresenceUpdates(ctx context.Context) {
	if pm.info == nil {
		pm.logger.Error("Cannot start presence updates: pod info not initialized")
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := pm.updatePresence(ctx); err != nil {
				pm.logger.Error("Failed to update presence", "error", err)
			}
		}
	}
}

// updatePresence updates the pod's last seen time and displays other pods
func (pm *PodManager) updatePresence(ctx context.Context) error {
	if pm.info == nil {
		return fmt.Errorf("pod info not initialized")
	}

	pm.info.LastSeen = time.Now()

	// Get existing pods
	pods, err := pm.getPods(ctx)
	if err != nil {
		return err
	}

	// Clean up dead pods
	pods = pm.cleanupDeadPods(ctx, pods)

	// Add or update current pod
	pods[pm.info.ID] = PodInfo{
		ID:        pm.info.ID,
		StartTime: pm.info.StartTime,
		LastSeen:  pm.info.LastSeen,
		Status:    pm.info.Status,
	}

	// Store updated pods
	if err := pm.registerPod(ctx); err != nil {
		return err
	}

	// Get current leader
	leaderID, err := pm.GetLeader(ctx)
	if err != nil {
		return err
	}

	// Clear line and print header
	fmt.Printf("\r\033[KActive Pods (%d): ", len(pods))

	// Print pod statuses
	first := true
	for id, info := range pods {
		if !first {
			fmt.Print(", ")
		}
		first = false

		status := "✓"
		if time.Since(info.LastSeen) > podTTL {
			status = "✗"
		}

		// Add leader and current pod indicators
		indicators := ""
		if id == leaderID {
			indicators += "👑" // Leader indicator
			if id == pm.info.ID {
				indicators += "⭐" // Current pod is leader
			}
		} else if id == pm.info.ID {
			indicators += "⚡" // Current pod (but not leader)
		}

		fmt.Printf("%s[%s]%s", status, id[:8], indicators)
	}
	fmt.Print("\n")

	return nil
}

// GetPodID returns the current pod's ID
func (pm *PodManager) GetPodID() string {
	if pm.info == nil {
		return ""
	}
	return pm.info.ID
}

// GetLeader returns the ID of the current leader pod
func (pm *PodManager) GetLeader(ctx context.Context) (string, error) {
	pods, err := pm.getPods(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pods: %w", err)
	}

	// Clean up dead pods
	pods = pm.cleanupDeadPods(ctx, pods)

	// If no pods are alive, return empty string
	if len(pods) == 0 {
		return "", nil
	}

	// Convert pods map to slice for sorting
	type podEntry struct {
		id        string
		startTime time.Time
	}
	podSlice := make([]podEntry, 0, len(pods))
	for id, info := range pods {
		podSlice = append(podSlice, podEntry{id: id, startTime: info.StartTime})
	}

	// Sort pods by start time
	sort.Slice(podSlice, func(i, j int) bool {
		return podSlice[i].startTime.Before(podSlice[j].startTime)
	})

	// Get the leader ID
	leaderID := podSlice[0].id

	// Update IsLeader status for all pods
	for id, info := range pods {
		info.IsLeader = (id == leaderID)
		pods[id] = info
	}

	// Store updated pods with leader status
	if err := pm.client.SetJSONWithExpiry(ctx, podRegistryKey, pods, 24*time.Hour); err != nil {
		return "", fmt.Errorf("failed to store updated pods: %w", err)
	}

	return leaderID, nil
}

// IsLeader checks if the current pod is the leader
func (pm *PodManager) IsLeader(ctx context.Context) (bool, error) {
	if pm.info == nil {
		return false, fmt.Errorf("pod info not initialized")
	}

	// Get pods to check leader status
	pods, err := pm.getPods(ctx)
	if err != nil {
		return false, err
	}

	// Get current pod info
	podInfo, exists := pods[pm.info.ID]
	if !exists {
		return false, fmt.Errorf("pod not found in registry")
	}

	return podInfo.IsLeader, nil
}

// GetLeader returns the ID of the current leader pod (global function)
func GetLeader() string {
	if instance == nil {
		return ""
	}
	leaderID, err := instance.GetLeader(context.Background())
	if err != nil {
		return ""
	}
	return leaderID
}

// IsLeader checks if the current pod is the leader (global function)
func IsLeader() bool {
	if instance == nil {
		instance.logger.Error("Pod manager instance is nil")
		return false
	}
	isLeader, err := instance.IsLeader(context.Background())
	if err != nil {
		instance.logger.Error("Failed to check leader status", "error", err)
		return false
	}
	return isLeader
}

// CheckPodHealth checks the health of all pods and updates their status
func (pm *PodManager) CheckPodHealth(ctx context.Context) error {
	// Get all pods from Redis
	pods, err := pm.getPods(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	// Check each pod's health
	for podID, pod := range pods {
		// Skip checking our own pod
		if podID == pm.info.ID {
			continue
		}

		// Check if pod is alive
		if time.Since(pod.LastSeen) > podTTL {
			pm.logger.Info("Pod is dead, removing from set", "pod_id", podID)

			// Remove pod from map
			delete(pods, podID)

			// Unassign all jobs from this pod
			if err := pm.assignment.UnassignJobsFromPod(ctx, podID); err != nil {
				pm.logger.Error("Failed to unassign jobs from pod", "pod_id", podID, "error", err)
			}
		}
	}

	// Store updated pods
	if err := pm.client.SetJSONWithExpiry(ctx, podRegistryKey, pods, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to store updated pods: %w", err)
	}

	// If we're the leader, assign jobs to available pods
	if IsLeader() {
		// Filter out our own pod from the list
		availablePods := make([]string, 0)
		for podID := range pods {
			if podID != pm.info.ID {
				availablePods = append(availablePods, podID)
			}
		}

		// Assign jobs to available pods
		if err := pm.assignment.AssignJobs(ctx, availablePods); err != nil {
			pm.logger.Error("Failed to assign jobs", "error", err)
		}
	} else {
		pm.logger.Info("Not the leader, skipping job assignment")
	}

	return nil
}
