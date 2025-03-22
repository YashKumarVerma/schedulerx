package scheduler

import (
	"fmt"
	"strings"
)

// ScheduleFetcher interface defines methods for retrieving command schedules
type ScheduleFetcher interface {
	// FetchSchedule retrieves the cron schedule and parameters for a given command
	// Returns:
	//   - string: cron schedule expression
	//   - []string: command parameters
	//   - error: any error that occurred
	FetchSchedule(commandID string) (string, []string, error)
}

// LocalScheduleFetcher implements ScheduleFetcher using local storage
type LocalScheduleFetcher struct {
	schedules map[string]CommandSchedule
}

// CommandSchedule represents the schedule and parameters for a command
type CommandSchedule struct {
	CronExpression string
	Parameters     []string
}

// NewLocalScheduleFetcher creates a new LocalScheduleFetcher instance with predefined schedules
func NewLocalScheduleFetcher() *LocalScheduleFetcher {
	fetcher := &LocalScheduleFetcher{
		schedules: make(map[string]CommandSchedule),
	}

	// Register predefined schedules
	fetcher.registerSchedules()

	return fetcher
}

// registerSchedules registers predefined command schedules
func (f *LocalScheduleFetcher) registerSchedules() {
	// Echo command - runs every 5 seconds
	f.schedules["echo"] = CommandSchedule{
		CronExpression: "*/5 * * * * *", // Every 5 seconds
		Parameters:     []string{"Heartbeat check"},
	}

	// List files - runs every minute
	f.schedules["ls"] = CommandSchedule{
		CronExpression: "* * * * *", // Every minute
		Parameters:     []string{"."},
	}

	// Disk usage - runs every 5 minutes
	f.schedules["du"] = CommandSchedule{
		CronExpression: "*/5 * * * *", // Every 5 minutes
		Parameters:     []string{"/"},
	}

	// Ping check - runs every 10 minutes
	f.schedules["ping"] = CommandSchedule{
		CronExpression: "*/10 * * * *", // Every 10 minutes
		Parameters:     []string{"google.com", "4", "1.0"},
	}

	// Shell command - runs every 30 minutes
	f.schedules["shell"] = CommandSchedule{
		CronExpression: "*/30 * * * *", // Every 30 minutes
		Parameters:     []string{"df -h"},
	}

	// Additional example schedules with different patterns
	f.schedules["hourly_check"] = CommandSchedule{
		CronExpression: "0 * * * *", // At minute 0 of every hour
		Parameters:     []string{"echo", "Hourly system check"},
	}

	f.schedules["daily_backup"] = CommandSchedule{
		CronExpression: "0 0 * * *", // At midnight every day
		Parameters:     []string{"echo", "Daily backup check"},
	}

	f.schedules["weekly_report"] = CommandSchedule{
		CronExpression: "0 0 * * 0", // At midnight on Sunday
		Parameters:     []string{"echo", "Weekly report generation"},
	}

	f.schedules["monthly_cleanup"] = CommandSchedule{
		CronExpression: "0 0 1 * *", // At midnight on the 1st of every month
		Parameters:     []string{"echo", "Monthly cleanup"},
	}

	// Complex patterns
	f.schedules["business_hours"] = CommandSchedule{
		CronExpression: "0 9-17 * * 1-5", // Every hour between 9 AM and 5 PM on weekdays
		Parameters:     []string{"echo", "Business hours check"},
	}

	f.schedules["quarterly"] = CommandSchedule{
		CronExpression: "0 0 1 */3 *", // At midnight on the 1st of every 3rd month
		Parameters:     []string{"echo", "Quarterly maintenance"},
	}

	// Multiple times per day
	f.schedules["multiple_daily"] = CommandSchedule{
		CronExpression: "0 8,12,18 * * *", // At 8 AM, 12 PM, and 6 PM every day
		Parameters:     []string{"echo", "Multiple daily check"},
	}

	// Every 2 hours
	f.schedules["bi_hourly"] = CommandSchedule{
		CronExpression: "0 */2 * * *", // Every 2 hours
		Parameters:     []string{"echo", "Bi-hourly check"},
	}
}

// FetchSchedule retrieves the schedule for a command from local storage
func (f *LocalScheduleFetcher) FetchSchedule(commandID string) (string, []string, error) {
	schedule, exists := f.schedules[commandID]
	if !exists {
		return "", nil, fmt.Errorf("no schedule found for command: %s", commandID)
	}

	return schedule.CronExpression, schedule.Parameters, nil
}

// ValidateCronExpression validates if the given string is a valid cron expression
func ValidateCronExpression(expr string) error {
	// Basic validation for cron expression format
	// Format: * * * * *
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return fmt.Errorf("invalid cron expression format: expected 5 fields, got %d", len(parts))
	}

	// Validate each field
	validFields := map[string]bool{
		"*": true,
		"?": true,
	}

	for i, part := range parts {
		// Allow wildcards and question marks
		if validFields[part] {
			continue
		}

		// Allow numbers and ranges
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return fmt.Errorf("invalid range format in field %d: %s", i+1, part)
			}
			continue
		}

		// Allow lists
		if strings.Contains(part, ",") {
			continue
		}

		// Allow step values
		if strings.Contains(part, "/") {
			continue
		}

		return fmt.Errorf("invalid character in field %d: %s", i+1, part)
	}

	return nil
}
