package command

import (
	"fmt"
	"os/exec"
	"strings"
)

// Command interface defines the methods that all commands must implement
type Command interface {
	// ID returns the unique identifier for the command
	ID() string
	// Description returns a human-readable description of the command
	Description() string
	// Execute runs the command with the given parameters
	Execute(params []string) error
	// Schedule returns the cron schedule and parameters for the command
	Schedule() (string, []string, error)
	// Parameters returns the default parameters for the command
	Parameters() []string
}

// CommandRegistry holds all available commands
type CommandRegistry struct {
	commands map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}
	registry.registerCommands()
	return registry
}

// registerCommands registers all available commands
func (r *CommandRegistry) registerCommands() {
	// Register echo command
	r.commands["echo"] = &EchoCommand{
		message: "",
	}

	// Register shell command
	r.commands["shell"] = &ShellCommand{
		command: "",
	}

	// Register ls command
	r.commands["ls"] = &ListFilesCommand{
		directory: ".",
	}

	// Register du command
	r.commands["du"] = &DiskUsageCommand{
		path: ".",
	}

	// Register ping command
	r.commands["ping"] = &PingCommand{
		host:     "localhost",
		count:    4,
		interval: 1.0,
	}
}

// GetCommand returns a command by its ID
func (r *CommandRegistry) GetCommand(id string) (Command, bool) {
	cmd, exists := r.commands[id]
	return cmd, exists
}

// GetCommandDescriptions used to log which commands are supported
func (r *CommandRegistry) GetCommandDescriptions() map[string]string {
	descriptions := make(map[string]string)
	for id, cmd := range r.commands {
		descriptions[id] = cmd.Description()
	}
	return descriptions
}

// GetCommands returns all registered commands
func (r *CommandRegistry) GetCommands() map[string]Command {
	return r.commands
}

// EchoCommand implements a simple echo command
type EchoCommand struct {
	message string
}

// NewEchoCommand creates a new EchoCommand
func NewEchoCommand(message string) *EchoCommand {
	return &EchoCommand{
		message: message,
	}
}

// ID returns the command identifier
func (c *EchoCommand) ID() string {
	return "echo"
}

// Description returns the command description
func (c *EchoCommand) Description() string {
	return "Echo a message to stdout"
}

// Execute runs the echo command
func (c *EchoCommand) Execute(params []string) error {
	if len(params) > 0 {
		fmt.Println(strings.Join(params, " "))
	} else {
		fmt.Println(c.message)
	}
	return nil
}

// Schedule returns the cron schedule and parameters for the command
func (c *EchoCommand) Schedule() (string, []string, error) {
	return "*/5 * * * * *", []string{"Heartbeat check"}, nil // Run every 5 seconds
}

// Parameters returns the default parameters for the command
func (c *EchoCommand) Parameters() []string {
	return []string{"Heartbeat check"}
}

// ShellCommand implements a shell command execution
type ShellCommand struct {
	command string
}

// NewShellCommand creates a new ShellCommand
func NewShellCommand(command string) *ShellCommand {
	return &ShellCommand{
		command: command,
	}
}

// ID returns the command identifier
func (c *ShellCommand) ID() string {
	return "shell"
}

// Description returns the command description
func (c *ShellCommand) Description() string {
	return "Execute a shell command"
}

// Execute runs the shell command
func (c *ShellCommand) Execute(params []string) error {
	cmd := exec.Command("sh", "-c", c.command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}
	fmt.Print(string(output))
	return nil
}

// Schedule returns the cron schedule and parameters for the command
func (c *ShellCommand) Schedule() (string, []string, error) {
	return "0 */30 * * * *", []string{"df -h"}, nil // Run every 30 minutes
}

// Parameters returns the default parameters for the command
func (c *ShellCommand) Parameters() []string {
	return []string{"df -h"}
}

// ListFilesCommand implements a directory listing command
type ListFilesCommand struct {
	directory string
}

// NewListFilesCommand creates a new ListFilesCommand
func NewListFilesCommand(directory string) *ListFilesCommand {
	return &ListFilesCommand{
		directory: directory,
	}
}

// ID returns the command identifier
func (c *ListFilesCommand) ID() string {
	return "ls"
}

// Description returns the command description
func (c *ListFilesCommand) Description() string {
	return "List files in a directory"
}

// Execute lists files in the specified directory
func (c *ListFilesCommand) Execute(params []string) error {
	dir := c.directory
	if len(params) > 0 {
		dir = params[0]
	}

	cmd := exec.Command("ls", "-la", dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list files: %w\nOutput: %s", err, string(output))
	}
	fmt.Print(string(output))
	return nil
}

// Schedule returns the cron schedule and parameters for the command
func (c *ListFilesCommand) Schedule() (string, []string, error) {
	return "0 * * * * *", []string{"."}, nil // Run every minute
}

// Parameters returns the default parameters for the command
func (c *ListFilesCommand) Parameters() []string {
	return []string{"."}
}

// DiskUsageCommand implements a disk usage command
type DiskUsageCommand struct {
	path string
}

// NewDiskUsageCommand creates a new DiskUsageCommand
func NewDiskUsageCommand(path string) *DiskUsageCommand {
	return &DiskUsageCommand{
		path: path,
	}
}

// ID returns the command identifier
func (c *DiskUsageCommand) ID() string {
	return "du"
}

// Description returns the command description
func (c *DiskUsageCommand) Description() string {
	return "Show disk usage for a path"
}

// Execute shows disk usage for the specified path
func (c *DiskUsageCommand) Execute(params []string) error {
	path := c.path
	if len(params) > 0 {
		path = params[0]
	}

	cmd := exec.Command("du", "-sh", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get disk usage: %w\nOutput: %s", err, string(output))
	}
	fmt.Print(string(output))
	return nil
}

// Schedule returns the cron schedule and parameters for the command
func (c *DiskUsageCommand) Schedule() (string, []string, error) {
	return "0 */5 * * * *", []string{"/"}, nil // Run every 5 minutes
}

// Parameters returns the default parameters for the command
func (c *DiskUsageCommand) Parameters() []string {
	return []string{"/"}
}

// PingCommand implements a network ping command
type PingCommand struct {
	host     string
	count    int
	interval float64
}

// NewPingCommand creates a new PingCommand
func NewPingCommand(host string, count int, interval float64) *PingCommand {
	return &PingCommand{
		host:     host,
		count:    count,
		interval: interval,
	}
}

// ID returns the command identifier
func (c *PingCommand) ID() string {
	return "ping"
}

// Description returns the command description
func (c *PingCommand) Description() string {
	return "Ping a host with specified count and interval"
}

// Execute runs the ping command
func (c *PingCommand) Execute(params []string) error {
	host := c.host
	if len(params) > 0 {
		host = params[0]
	}

	args := []string{
		"-c", fmt.Sprintf("%d", c.count),
		"-i", fmt.Sprintf("%f", c.interval),
		host,
	}

	cmd := exec.Command("ping", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ping failed: %w\nOutput: %s", err, string(output))
	}
	fmt.Print(string(output))
	return nil
}

// Schedule returns the cron schedule and parameters for the command
func (c *PingCommand) Schedule() (string, []string, error) {
	return "0 */10 * * * *", []string{"google.com", "4", "1.0"}, nil // Run every 10 minutes
}

// Parameters returns the default parameters for the command
func (c *PingCommand) Parameters() []string {
	return []string{"google.com", "4", "1.0"}
}
