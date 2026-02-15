package handlers

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateRunRequest represents a run creation request
type CreateRunRequest struct {
	// Workflow identification
	Flow   string            `json:"flow"`   // Flow workflow name
	Module string            `json:"module"` // Module workflow name
	Target string            `json:"target,omitempty"`
	Params map[string]string `json:"params"`

	// Multi-target support
	Targets    []string `json:"targets,omitempty"`     // Array of targets to run against
	TargetFile string   `json:"target_file,omitempty"` // Path to file containing targets (one per line)

	// Concurrency control
	Concurrency int `json:"concurrency,omitempty"` // Number of concurrent runs (default: 1)

	// Priority and timeout
	Priority string `json:"priority,omitempty"` // low, normal, high, critical (default: high)
	RunMode  string `json:"run_mode,omitempty"` // local, distributed, cloud (default: local)
	Timeout  int    `json:"timeout,omitempty"`  // Timeout in minutes (0 = no timeout)

	// Runner configuration
	RunnerType  string `json:"runner_type,omitempty"`  // host, docker, ssh (default: host)
	DockerImage string `json:"docker_image,omitempty"` // Docker image to use when runner_type=docker
	SSHHost     string `json:"ssh_host,omitempty"`     // SSH host when runner_type=ssh

	// Scheduling options
	Schedule         string `json:"schedule,omitempty"`           // Cron expression for scheduled scans
	ScheduleEnabled  bool   `json:"schedule_enabled,omitempty"`   // Enable scheduled execution
	NotifyOnComplete bool   `json:"notify_on_complete,omitempty"` // Send notification when run completes

	// Execution options (mirrors CLI flags)
	ThreadsHold     int    `json:"threads_hold,omitempty"`     // Override thread count (0 = use tactic default)
	EmptyTarget     bool   `json:"empty_target,omitempty"`     // Run without target (generates placeholder target)
	Repeat          bool   `json:"repeat,omitempty"`           // Repeat run after completion
	RepeatWaitTime  string `json:"repeat_wait_time,omitempty"` // Wait time between repeats (e.g., 30s, 20m, 10h, 1d)
	HeuristicsCheck string `json:"heuristics_check,omitempty"` // Heuristics check level: none, basic, advanced

	// Validation
	SkipValidation bool `json:"skip_validation,omitempty"` // Skip input validation for dangerous shell characters
}

// CreateScheduleRequest represents a schedule creation request
type CreateScheduleRequest struct {
	Name         string                 `json:"name"`
	WorkflowName string                 `json:"workflow_name"`
	WorkflowKind string                 `json:"workflow_kind,omitempty"` // "module" or "flow"
	Target       string                 `json:"target,omitempty"`
	Workspace    string                 `json:"workspace,omitempty"`
	Schedule     string                 `json:"schedule"` // cron expression
	Params       map[string]interface{} `json:"params,omitempty"`
	EventTopic   string                 `json:"event_topic,omitempty"`
	WatchPath    string                 `json:"watch_path,omitempty"`
	TriggerType  string                 `json:"trigger_type,omitempty"` // cron, event, watch, manual
	Enabled      bool                   `json:"enabled"`
	RunnerType   string                 `json:"runner_type,omitempty"`
}

// UpdateScheduleRequest represents a schedule update request
type UpdateScheduleRequest struct {
	Name      string                 `json:"name,omitempty"`
	Target    string                 `json:"target,omitempty"`
	Workspace string                 `json:"workspace,omitempty"`
	Schedule  string                 `json:"schedule,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Enabled   *bool                  `json:"enabled,omitempty"`
}

// BinaryStatusEntry represents a binary with its registry info and installation status
type BinaryStatusEntry struct {
	Desc                string            `json:"desc,omitempty"`
	RepoLink            string            `json:"repo_link,omitempty"`
	Version             string            `json:"version,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	ValidateCommand     string            `json:"valide-command,omitempty"`
	Linux               map[string]string `json:"linux,omitempty"`
	Darwin              map[string]string `json:"darwin,omitempty"`
	Windows             map[string]string `json:"windows,omitempty"`
	CommandLinux        map[string]string `json:"command-linux,omitempty"`
	CommandDarwin       map[string]string `json:"command-darwin,omitempty"`
	CommandDual         map[string]string `json:"command-dual,omitempty"`
	MultiCommandsLinux  []string          `json:"multi-commands-linux,omitempty"`
	MultiCommandsDarwin []string          `json:"multi-commands-darwin,omitempty"`
	Installed           bool              `json:"installed"`
	Path                string            `json:"path,omitempty"`
	Optional            bool              `json:"optional"`
}

// InstallRequest represents an installation request
type InstallRequest struct {
	Type            string   `json:"type"`                       // "binary" or "workflow"
	Names           []string `json:"names,omitempty"`            // Binary names to install (for type=binary)
	Source          string   `json:"source,omitempty"`           // Git URL, zip URL, or file path (for type=workflow)
	RegistryURL     string   `json:"registry_url,omitempty"`     // Custom registry URL (optional, for type=binary)
	InstallAll      bool     `json:"install_all,omitempty"`      // Install all binaries from registry (for type=binary)
	InstallOptional bool     `json:"install_optional,omitempty"` // Include optional binaries (only with install_all)
	RegistryMode    string   `json:"registry_mode,omitempty"`    // "direct-fetch" or "nix-build" (default: direct-fetch)
}

// FunctionEvalRequest represents a function evaluation request
type FunctionEvalRequest struct {
	Script string            `json:"script"`
	Target string            `json:"target,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// FunctionListResponse represents a function in the list response
type FunctionListResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ReturnType  string   `json:"return_type"`
	Example     string   `json:"example,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// readTargetsFromFile reads targets from a file (one per line)
func readTargetsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result, scanner.Err()
}

// dangerousChars contains shell metacharacters that could enable command injection
const dangerousChars = " ;'\"|&`$(){}[]"

// containsDangerousChars checks if a string contains any dangerous shell metacharacters
func containsDangerousChars(s string) bool {
	return strings.ContainsAny(s, dangerousChars)
}

// validateCreateRunInput validates all string fields in a CreateRunRequest for dangerous characters.
// Returns an error describing which field failed and what character was found.
func validateCreateRunInput(req *CreateRunRequest) error {
	type fieldCheck struct {
		name  string
		value string
	}

	fields := []fieldCheck{
		{"target", req.Target},
		{"target_file", req.TargetFile},
		{"flow", req.Flow},
		{"module", req.Module},
		{"ssh_host", req.SSHHost},
		{"docker_image", req.DockerImage},
		{"repeat_wait_time", req.RepeatWaitTime},
	}

	for _, f := range fields {
		if f.value == "" {
			continue
		}
		for _, ch := range f.value {
			if strings.ContainsRune(dangerousChars, ch) {
				return fmt.Errorf("field %q contains forbidden character %q", f.name, string(ch))
			}
		}
	}

	for i, t := range req.Targets {
		for _, ch := range t {
			if strings.ContainsRune(dangerousChars, ch) {
				return fmt.Errorf("targets[%d] contains forbidden character %q", i, string(ch))
			}
		}
	}

	for k, v := range req.Params {
		for _, ch := range v {
			if strings.ContainsRune(dangerousChars, ch) {
				return fmt.Errorf("params[%q] contains forbidden character %q", k, string(ch))
			}
		}
	}

	return nil
}

// deduplicateTargets removes duplicates while preserving order
func deduplicateTargets(targets []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	return result
}
