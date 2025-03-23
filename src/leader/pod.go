package leader

import (
	"time"
)

const (
	// Redis keys
	PodSetKey     = "schedulerx:pods"
	PodDetailsKey = "schedulerx:pod:%s" // Format string for pod details
	PodTimeout    = 10 * time.Second    // Time after which a pod is considered dead
)

// Pod represents information about a running pod
type Pod struct {
	ID            string    `json:"id"`
	StartTime     time.Time `json:"start_time"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Status        string    `json:"status"`
}
