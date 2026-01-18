package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StepTimeout string

func (t *StepTimeout) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var i int
	if err := unmarshal(&i); err == nil {
		if i < 0 {
			i = 0
		}
		*t = StepTimeout(strconv.Itoa(i))
		return nil
	}

	var s string
	if err := unmarshal(&s); err == nil {
		*t = StepTimeout(strings.TrimSpace(s))
		return nil
	}

	return fmt.Errorf("invalid timeout")
}

func (t StepTimeout) MarshalYAML() (interface{}, error) {
	s := strings.TrimSpace(string(t))
	if s == "" {
		return nil, nil
	}

	if isDigits(s) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return s, nil
		}
		return i, nil
	}

	return s, nil
}

func (t StepTimeout) Duration() (time.Duration, error) {
	s := strings.TrimSpace(string(t))
	if s == "" {
		return 0, nil
	}

	if isDigits(s) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid timeout: %w", err)
		}
		if i <= 0 {
			return 0, nil
		}
		return time.Duration(i) * time.Second, nil
	}

	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		if !isDigits(daysStr) {
			return 0, fmt.Errorf("invalid timeout: %s", s)
		}
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, fmt.Errorf("invalid timeout: %w", err)
		}
		if days <= 0 {
			return 0, nil
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout: %w", err)
	}
	if d <= 0 {
		return 0, nil
	}
	return d, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

type StepThreads string

func (t *StepThreads) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var i int
	if err := unmarshal(&i); err == nil {
		if i < 0 {
			i = 0
		}
		*t = StepThreads(strconv.Itoa(i))
		return nil
	}

	var s string
	if err := unmarshal(&s); err == nil {
		*t = StepThreads(strings.TrimSpace(s))
		return nil
	}

	return fmt.Errorf("invalid threads")
}

func (t StepThreads) MarshalYAML() (interface{}, error) {
	s := strings.TrimSpace(string(t))
	if s == "" {
		return nil, nil
	}

	if isDigits(s) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return s, nil
		}
		return i, nil
	}

	return s, nil
}

func (t StepThreads) Int() (int, error) {
	s := strings.TrimSpace(string(t))
	if s == "" {
		return 0, nil
	}

	i, err := strconv.Atoi(s)
	if err == nil {
		if i <= 0 {
			return 0, nil
		}
		return i, nil
	}

	f, ferr := strconv.ParseFloat(s, 64)
	if ferr != nil {
		return 0, fmt.Errorf("invalid threads: %w", err)
	}

	if f <= 0 {
		return 0, nil
	}

	if f != float64(int(f)) {
		return 0, fmt.Errorf("invalid threads: %s", s)
	}

	return int(f), nil
}

// StepRunnerConfig holds per-step runner configuration for remote-bash steps
// The runner type is specified separately in Step.StepRunner
type StepRunnerConfig struct {
	*RunnerConfig `yaml:",inline"` // Embed all RunnerConfig fields (image, host, etc.)
}

// Step represents a single execution step in a module
type Step struct {
	Name         string      `yaml:"name"`
	Type         StepType    `yaml:"type"`
	StepRunner   RunnerType  `yaml:"step_runner"` // Runner for this step: local (default), docker, ssh
	PreCondition string      `yaml:"pre_condition"`
	Log          string      `yaml:"log"`
	Timeout      StepTimeout `yaml:"timeout,omitempty"`

	// Bash step fields
	Command          string   `yaml:"command"`
	Commands         []string `yaml:"commands"`
	ParallelCommands []string `yaml:"parallel_commands"`
	StdFile          string   `yaml:"std_file"` // File path to save stdout/stderr output

	// Structured argument fields (for bash/remote-bash steps)
	// These are templated and joined with Command in order: command + speed + config + input + output
	SpeedArgs  string `yaml:"speed_args"`
	ConfigArgs string `yaml:"config_args"`
	InputArgs  string `yaml:"input_args"`
	OutputArgs string `yaml:"output_args"`

	// Function step fields
	Function          string   `yaml:"function"`
	Functions         []string `yaml:"functions"`
	ParallelFunctions []string `yaml:"parallel_functions"`

	// Parallel step fields
	ParallelSteps []Step `yaml:"parallel_steps"`

	// Foreach step fields
	Input    string      `yaml:"input"`
	Variable string      `yaml:"variable"`
	Threads  StepThreads `yaml:"threads,omitempty"`
	Step     *Step       `yaml:"step"`

	// Remote-bash step fields
	StepRunnerConfig *StepRunnerConfig `yaml:"step_runner_config"`
	StepRemoteFile   string            `yaml:"step_remote_file"` // File path on remote (Docker/SSH) to copy after execution
	HostOutputFile   string            `yaml:"host_output_file"` // Local path to copy the remote file to

	// HTTP step fields
	URL         string            `yaml:"url"`
	Method      string            `yaml:"method"`
	Headers     map[string]string `yaml:"headers"`
	RequestBody string            `yaml:"request_body"`

	// LLM step fields
	Messages       []LLMMessage           `yaml:"messages"`
	Tools          []LLMTool              `yaml:"tools,omitempty"`
	ToolChoice     interface{}            `yaml:"tool_choice,omitempty"`
	LLMConfig      *LLMStepConfig         `yaml:"llm_config,omitempty"`
	IsEmbedding    bool                   `yaml:"is_embedding,omitempty"`
	EmbeddingInput []string               `yaml:"embedding_input,omitempty"`
	ExtraLLMParams map[string]interface{} `yaml:"extra_llm_parameters,omitempty"`

	// Common fields
	Exports   map[string]string `yaml:"exports"`
	OnSuccess []Action          `yaml:"on_success"`
	OnError   []Action          `yaml:"on_error"`
	Decision  *DecisionConfig   `yaml:"decision,omitempty"`
}

// DecisionCase represents a single case in switch-style decision
type DecisionCase struct {
	Goto string `yaml:"goto"`
}

// DecisionConfig supports switch/case routing for conditional workflow branching.
//
// Switch/case syntax:
//
//	decision:
//	  switch: "{{variable}}"
//	  cases:
//	    "value1": { goto: step-a }
//	    "value2": { goto: step-b }
//	  default:
//	    goto: fallback-step
type DecisionConfig struct {
	Switch  string                  `yaml:"switch,omitempty"`
	Cases   map[string]DecisionCase `yaml:"cases,omitempty"`
	Default *DecisionCase           `yaml:"default,omitempty"`
}

// Action represents on_success/on_error handler
type Action struct {
	Action    ActionType        `yaml:"action"`
	Message   string            `yaml:"message"`
	Condition string            `yaml:"condition"`
	Name      string            `yaml:"name"`      // for export action
	Value     interface{}       `yaml:"value"`     // for export action
	Type      StepType          `yaml:"type"`      // for run action
	Command   string            `yaml:"command"`   // for run bash action
	Functions []string          `yaml:"functions"` // for run function action
	Export    map[string]string `yaml:"export"`    // for run function action
	Notify    string            `yaml:"notify"`    // notification message
}

// IsBashStep returns true if this is a bash step
func (s *Step) IsBashStep() bool {
	return s.Type == StepTypeBash
}

// IsFunctionStep returns true if this is a function step
func (s *Step) IsFunctionStep() bool {
	return s.Type == StepTypeFunction
}

// IsParallelStep returns true if this is a parallel step
func (s *Step) IsParallelStep() bool {
	return s.Type == StepTypeParallel
}

// IsForeachStep returns true if this is a foreach step
func (s *Step) IsForeachStep() bool {
	return s.Type == StepTypeForeach
}

// IsRemoteBashStep returns true if this is a remote-bash step
func (s *Step) IsRemoteBashStep() bool {
	return s.Type == StepTypeRemoteBash
}

// IsHTTPStep returns true if this is an HTTP step
func (s *Step) IsHTTPStep() bool {
	return s.Type == StepTypeHTTP
}

// IsLLMStep returns true if this is an LLM step
func (s *Step) IsLLMStep() bool {
	return s.Type == StepTypeLLM
}

// GetStepRunner returns the step runner type, defaulting to host/local
func (s *Step) GetStepRunner() RunnerType {
	if s.StepRunner == "" {
		return RunnerTypeHost // default to local
	}
	return s.StepRunner
}

// HasParallelCommands returns true if step has parallel commands
func (s *Step) HasParallelCommands() bool {
	return len(s.ParallelCommands) > 0
}

// HasParallelFunctions returns true if step has parallel functions
func (s *Step) HasParallelFunctions() bool {
	return len(s.ParallelFunctions) > 0
}

// HasDecision returns true if step has decision routing
func (s *Step) HasDecision() bool {
	if s.Decision == nil {
		return false
	}
	return s.Decision.Switch != "" || len(s.Decision.Cases) > 0
}

// HasExports returns true if step exports variables
func (s *Step) HasExports() bool {
	return len(s.Exports) > 0
}

// GetCommands returns the list of commands to execute
// Returns single command as slice if Commands is empty
func (s *Step) GetCommands() []string {
	if len(s.Commands) > 0 {
		return s.Commands
	}
	if s.Command != "" {
		return []string{s.Command}
	}
	return nil
}

// GetFunctions returns the list of functions to execute
// Returns single function as slice if Functions is empty
func (s *Step) GetFunctions() []string {
	if len(s.Functions) > 0 {
		return s.Functions
	}
	if s.Function != "" {
		return []string{s.Function}
	}
	return nil
}

// Clone creates a shallow copy of the step with new slices for Commands
func (s *Step) Clone() *Step {
	cloned := *s

	// Deep copy slices to avoid modifying originals
	if len(s.Commands) > 0 {
		cloned.Commands = make([]string, len(s.Commands))
		copy(cloned.Commands, s.Commands)
	}
	if len(s.ParallelCommands) > 0 {
		cloned.ParallelCommands = make([]string, len(s.ParallelCommands))
		copy(cloned.ParallelCommands, s.ParallelCommands)
	}
	if len(s.Functions) > 0 {
		cloned.Functions = make([]string, len(s.Functions))
		copy(cloned.Functions, s.Functions)
	}
	if len(s.ParallelFunctions) > 0 {
		cloned.ParallelFunctions = make([]string, len(s.ParallelFunctions))
		copy(cloned.ParallelFunctions, s.ParallelFunctions)
	}

	// Deep copy StepRunnerConfig
	if s.StepRunnerConfig != nil {
		clonedConfig := &StepRunnerConfig{}
		if s.StepRunnerConfig.RunnerConfig != nil {
			cfg := *s.StepRunnerConfig.RunnerConfig
			// Deep copy slices in RunnerConfig
			if len(cfg.Volumes) > 0 {
				cfg.Volumes = make([]string, len(s.StepRunnerConfig.Volumes))
				copy(cfg.Volumes, s.StepRunnerConfig.Volumes)
			}
			if len(cfg.Env) > 0 {
				cfg.Env = make(map[string]string, len(s.StepRunnerConfig.Env))
				for k, v := range s.StepRunnerConfig.Env {
					cfg.Env[k] = v
				}
			}
			clonedConfig.RunnerConfig = &cfg
		}
		cloned.StepRunnerConfig = clonedConfig
	}

	// Deep copy HTTP Headers map
	if len(s.Headers) > 0 {
		cloned.Headers = make(map[string]string, len(s.Headers))
		for k, v := range s.Headers {
			cloned.Headers[k] = v
		}
	}

	// Deep copy LLM fields
	if len(s.Messages) > 0 {
		cloned.Messages = make([]LLMMessage, len(s.Messages))
		copy(cloned.Messages, s.Messages)
	}
	if len(s.Tools) > 0 {
		cloned.Tools = make([]LLMTool, len(s.Tools))
		copy(cloned.Tools, s.Tools)
	}
	if len(s.EmbeddingInput) > 0 {
		cloned.EmbeddingInput = make([]string, len(s.EmbeddingInput))
		copy(cloned.EmbeddingInput, s.EmbeddingInput)
	}
	if len(s.ExtraLLMParams) > 0 {
		cloned.ExtraLLMParams = make(map[string]interface{}, len(s.ExtraLLMParams))
		for k, v := range s.ExtraLLMParams {
			cloned.ExtraLLMParams[k] = v
		}
	}

	return &cloned
}
