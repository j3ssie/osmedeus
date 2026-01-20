package executor

import (
	"context"
	"fmt"
	"regexp"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/runner"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// functionCallPattern matches function call syntax like functionName(...)
var functionCallPattern = regexp.MustCompile(`\w+\s*\(`)

// StepDispatcher dispatches steps to appropriate executors
type StepDispatcher struct {
	registry         *PluginRegistry
	templateEngine   template.TemplateEngine
	batchRenderer    template.BatchRenderer // For optimized batch rendering
	functionRegistry *functions.Registry
	dryRun           bool
	runner           runner.Runner
	enableBatch      bool // Enable batch template rendering
	// Keep direct references to executors that need special configuration
	bashExecutor *BashExecutor
	llmExecutor  *LLMExecutor
}

// SetDryRun enables or disables dry-run mode for the dispatcher
func (d *StepDispatcher) SetDryRun(dryRun bool) {
	d.dryRun = dryRun
}

// SetSilent enables or disables silent mode for executors that support it
func (d *StepDispatcher) SetSilent(silent bool) {
	d.llmExecutor.SetSilent(silent)
}

// SetRunner sets the runner for command execution
func (d *StepDispatcher) SetRunner(r runner.Runner) {
	d.runner = r
	d.bashExecutor.SetRunner(r)
}

// StepDispatcherConfig holds configuration for the step dispatcher
type StepDispatcherConfig struct {
	UseShardedEngine bool // Use sharded template engine for better concurrency
	EnableBatch      bool // Enable batch template rendering
	ShardCount       int  // Number of shards (default: 16)
	ShardCacheSize   int  // Cache size per shard (default: 64)
}

// DefaultStepDispatcherConfig returns the default configuration
func DefaultStepDispatcherConfig() StepDispatcherConfig {
	return StepDispatcherConfig{
		UseShardedEngine: true, // Default to sharded engine for better performance
		EnableBatch:      true, // Default to batch rendering
		ShardCount:       16,
		ShardCacheSize:   64,
	}
}

// NewStepDispatcher creates a new step dispatcher with default configuration
func NewStepDispatcher() *StepDispatcher {
	return NewStepDispatcherWithConfig(DefaultStepDispatcherConfig())
}

// NewStepDispatcherWithConfig creates a new step dispatcher with custom configuration
func NewStepDispatcherWithConfig(cfg StepDispatcherConfig) *StepDispatcher {
	var engine template.TemplateEngine
	var batchRenderer template.BatchRenderer

	if cfg.UseShardedEngine {
		shardedCfg := template.ShardedEngineConfig{
			ShardCount:     cfg.ShardCount,
			ShardCacheSize: cfg.ShardCacheSize,
			EnablePooling:  true,
		}
		shardedEngine := template.NewShardedEngineWithConfig(shardedCfg)
		engine = shardedEngine
		batchRenderer = shardedEngine
	} else {
		stdEngine := template.NewEngine()
		engine = stdEngine
		batchRenderer = stdEngine
	}

	d := &StepDispatcher{
		registry:         NewPluginRegistry(),
		templateEngine:   engine,
		batchRenderer:    batchRenderer,
		functionRegistry: functions.NewRegistry(),
		enableBatch:      cfg.EnableBatch,
	}

	// Create executors
	d.bashExecutor = NewBashExecutor(engine)
	d.llmExecutor = NewLLMExecutor(engine)

	// Register all built-in plugins
	d.registry.Register(d.bashExecutor)
	d.registry.Register(NewFunctionExecutor(engine, d.functionRegistry))
	d.registry.Register(NewParallelExecutor(d))
	d.registry.Register(NewForeachExecutor(d, engine))
	d.registry.Register(NewRemoteBashExecutor(engine))
	d.registry.Register(NewHTTPExecutor(engine))
	d.registry.Register(d.llmExecutor)

	return d
}

// RegisterPlugin allows external plugin registration
func (d *StepDispatcher) RegisterPlugin(plugin StepExecutorPlugin) {
	d.registry.Register(plugin)
}

// SetConfig passes config to executors that need it
func (d *StepDispatcher) SetConfig(cfg *config.Config) {
	d.llmExecutor.SetConfig(cfg)
}

// Dispatch dispatches a step to the appropriate executor
func (d *StepDispatcher) Dispatch(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	log := logger.Get()

	log.Debug("Dispatching step",
		zap.String("step_name", step.Name),
		zap.String("step_type", string(step.Type)),
		zap.Bool("dry_run", d.dryRun),
	)

	// Render templates in step fields
	log.Debug("Rendering step templates")
	renderedStep, err := d.renderStep(step, execCtx)
	if err != nil {
		log.Debug("Template rendering failed", zap.Error(err))
		return nil, fmt.Errorf("template rendering failed: %w", err)
	}

	log.Debug("Step templates rendered",
		zap.String("command", renderedStep.Command),
	)

	// Log step message if provided
	if renderedStep.Log != "" {
		log.Info(renderedStep.Log,
			zap.String("step", step.Name),
		)
	}

	// Dispatch based on step type using plugin registry
	log.Debug("Dispatching to executor",
		zap.String("executor_type", string(step.Type)),
	)

	plugin, ok := d.registry.Get(step.Type)
	if !ok {
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}

	log.Debug("Using plugin", zap.String("plugin_name", plugin.Name()))
	result, err := plugin.Execute(ctx, renderedStep, execCtx)

	if err != nil {
		log.Debug("Step execution failed",
			zap.String("step", step.Name),
			zap.Error(err),
		)
		return result, err
	}

	log.Debug("Step execution completed",
		zap.String("step", step.Name),
		zap.String("status", string(result.Status)),
	)

	// Process exports
	if step.HasExports() {
		log.Debug("Processing exports",
			zap.Int("export_count", len(step.Exports)),
		)

		// Merge auto-exports (e.g., from HTTP steps) into vars before evaluating user exports
		vars := execCtx.GetVariables()
		if result.Exports != nil {
			for k, v := range result.Exports {
				vars[k] = v
			}
		}

		// Render template variables in export values first, then evaluate if needed
		exports := make(map[string]interface{}, len(step.Exports))
		for name, expr := range step.Exports {
			rendered, err := d.templateEngine.Render(expr, vars)
			if err != nil {
				log.Warn("Failed to render export value, using original",
					zap.String("export", name),
					zap.Error(err))
				rendered = expr
			}

			// Only evaluate with JS if the rendered value contains a function call
			// Otherwise, use the rendered string directly
			if functionCallPattern.MatchString(rendered) {
				value, err := d.functionRegistry.Execute(rendered, vars)
				if err != nil {
					return result, fmt.Errorf("export evaluation failed for %s: %w", name, err)
				}
				exports[name] = value
			} else {
				// Use rendered value directly as a string
				exports[name] = rendered
			}
		}
		if result.Exports == nil {
			result.Exports = make(map[string]interface{})
		}
		for k, v := range exports {
			result.Exports[k] = v
		}
		log.Debug("Exports processed", zap.Int("total_exports", len(result.Exports)))
	}

	return result, nil
}

// collectRenderRequests gathers all non-empty template strings from a step
func collectRenderRequests(step *core.Step) []template.RenderRequest {
	var requests []template.RenderRequest
	add := func(key, tmpl string) {
		if tmpl != "" {
			requests = append(requests, template.RenderRequest{Key: key, Template: tmpl})
		}
	}

	// String fields
	add("Command", step.Command)
	add("SpeedArgs", step.SpeedArgs)
	add("ConfigArgs", step.ConfigArgs)
	add("InputArgs", step.InputArgs)
	add("OutputArgs", step.OutputArgs)
	add("StdFile", step.StdFile)
	add("Function", step.Function)
	add("Input", step.Input)
	add("Log", step.Log)
	add("Timeout", string(step.Timeout))
	add("Threads", string(step.Threads))
	add("URL", step.URL)
	add("Method", step.Method)
	add("RequestBody", step.RequestBody)
	add("StepRunner", string(step.StepRunner))
	add("StepRemoteFile", step.StepRemoteFile)
	add("HostOutputFile", step.HostOutputFile)

	// Slice fields
	for i, cmd := range step.Commands {
		add(fmt.Sprintf("Commands[%d]", i), cmd)
	}
	for i, cmd := range step.ParallelCommands {
		add(fmt.Sprintf("ParallelCommands[%d]", i), cmd)
	}
	for i, fn := range step.Functions {
		add(fmt.Sprintf("Functions[%d]", i), fn)
	}
	for i, fn := range step.ParallelFunctions {
		add(fmt.Sprintf("ParallelFunctions[%d]", i), fn)
	}
	for i, input := range step.EmbeddingInput {
		add(fmt.Sprintf("EmbeddingInput[%d]", i), input)
	}

	// Map fields
	for k, v := range step.Headers {
		add(fmt.Sprintf("Headers[%s]", k), v)
	}

	// RunnerConfig fields
	if step.StepRunnerConfig != nil && step.StepRunnerConfig.RunnerConfig != nil {
		cfg := step.StepRunnerConfig.RunnerConfig
		add("RunnerConfig.Image", cfg.Image)
		add("RunnerConfig.Host", cfg.Host)
		add("RunnerConfig.User", cfg.User)
		add("RunnerConfig.Password", cfg.Password)
		add("RunnerConfig.KeyFile", cfg.KeyFile)
		add("RunnerConfig.WorkDir", cfg.WorkDir)
		add("RunnerConfig.Network", cfg.Network)
		for k, v := range cfg.Env {
			add(fmt.Sprintf("RunnerConfig.Env[%s]", k), v)
		}
		for i, v := range cfg.Volumes {
			add(fmt.Sprintf("RunnerConfig.Volumes[%d]", i), v)
		}
	}

	return requests
}

// renderStep renders all template fields in a step
func (d *StepDispatcher) renderStep(step *core.Step, execCtx *core.ExecutionContext) (*core.Step, error) {
	vars := execCtx.GetVariables()

	// Use batch rendering if enabled and available
	if d.enableBatch && d.batchRenderer != nil {
		return d.renderStepBatch(step, vars)
	}

	return d.renderStepSequential(step, vars)
}

// renderStepBatch renders step templates using batch mode for fewer lock acquisitions
func (d *StepDispatcher) renderStepBatch(step *core.Step, vars map[string]any) (*core.Step, error) {
	requests := collectRenderRequests(step)
	if len(requests) == 0 {
		// No templates to render, but still need to handle LLM messages
		if len(step.Messages) > 0 {
			rendered := *step
			if err := d.renderLLMMessages(&rendered, vars); err != nil {
				return nil, err
			}
			return &rendered, nil
		}
		return step, nil
	}

	results, err := d.batchRenderer.RenderBatch(requests, vars)
	if err != nil {
		return nil, err
	}

	rendered := *step
	get := func(key string) string { return results[key] }

	// Apply results to string fields
	if v := get("Command"); v != "" {
		rendered.Command = v
	}
	if v := get("SpeedArgs"); v != "" {
		rendered.SpeedArgs = v
	}
	if v := get("ConfigArgs"); v != "" {
		rendered.ConfigArgs = v
	}
	if v := get("InputArgs"); v != "" {
		rendered.InputArgs = v
	}
	if v := get("OutputArgs"); v != "" {
		rendered.OutputArgs = v
	}
	if v := get("StdFile"); v != "" {
		rendered.StdFile = v
	}
	if v := get("Function"); v != "" {
		rendered.Function = v
	}
	if v := get("Input"); v != "" {
		rendered.Input = v
	}
	if v := get("Log"); v != "" {
		rendered.Log = v
	}
	if v := get("Timeout"); v != "" {
		rendered.Timeout = core.StepTimeout(v)
	}
	if v := get("Threads"); v != "" {
		rendered.Threads = core.StepThreads(v)
	}
	if v := get("URL"); v != "" {
		rendered.URL = v
	}
	if v := get("Method"); v != "" {
		rendered.Method = v
	}
	if v := get("RequestBody"); v != "" {
		rendered.RequestBody = v
	}
	if v := get("StepRunner"); v != "" {
		rendered.StepRunner = core.RunnerType(v)
	}
	if v := get("StepRemoteFile"); v != "" {
		rendered.StepRemoteFile = v
	}
	if v := get("HostOutputFile"); v != "" {
		rendered.HostOutputFile = v
	}

	// Apply results to slice fields
	if len(step.Commands) > 0 {
		rendered.Commands = make([]string, len(step.Commands))
		for i := range step.Commands {
			rendered.Commands[i] = get(fmt.Sprintf("Commands[%d]", i))
		}
	}
	if len(step.ParallelCommands) > 0 {
		rendered.ParallelCommands = make([]string, len(step.ParallelCommands))
		for i := range step.ParallelCommands {
			rendered.ParallelCommands[i] = get(fmt.Sprintf("ParallelCommands[%d]", i))
		}
	}
	if len(step.Functions) > 0 {
		rendered.Functions = make([]string, len(step.Functions))
		for i := range step.Functions {
			rendered.Functions[i] = get(fmt.Sprintf("Functions[%d]", i))
		}
	}
	if len(step.ParallelFunctions) > 0 {
		rendered.ParallelFunctions = make([]string, len(step.ParallelFunctions))
		for i := range step.ParallelFunctions {
			rendered.ParallelFunctions[i] = get(fmt.Sprintf("ParallelFunctions[%d]", i))
		}
	}
	if len(step.EmbeddingInput) > 0 {
		rendered.EmbeddingInput = make([]string, len(step.EmbeddingInput))
		for i := range step.EmbeddingInput {
			rendered.EmbeddingInput[i] = get(fmt.Sprintf("EmbeddingInput[%d]", i))
		}
	}

	// Apply results to map fields
	if len(step.Headers) > 0 {
		rendered.Headers = make(map[string]string, len(step.Headers))
		for k := range step.Headers {
			rendered.Headers[k] = get(fmt.Sprintf("Headers[%s]", k))
		}
	}

	// Apply results to RunnerConfig
	if step.StepRunnerConfig != nil && step.StepRunnerConfig.RunnerConfig != nil {
		cfg := *step.StepRunnerConfig.RunnerConfig
		if v := get("RunnerConfig.Image"); v != "" {
			cfg.Image = v
		}
		if v := get("RunnerConfig.Host"); v != "" {
			cfg.Host = v
		}
		if v := get("RunnerConfig.User"); v != "" {
			cfg.User = v
		}
		if v := get("RunnerConfig.Password"); v != "" {
			cfg.Password = v
		}
		if v := get("RunnerConfig.KeyFile"); v != "" {
			cfg.KeyFile = v
		}
		if v := get("RunnerConfig.WorkDir"); v != "" {
			cfg.WorkDir = v
		}
		if v := get("RunnerConfig.Network"); v != "" {
			cfg.Network = v
		}
		// Apply Env map
		if len(step.StepRunnerConfig.Env) > 0 {
			cfg.Env = make(map[string]string, len(step.StepRunnerConfig.Env))
			for k := range step.StepRunnerConfig.Env {
				cfg.Env[k] = get(fmt.Sprintf("RunnerConfig.Env[%s]", k))
			}
		}
		// Apply Volumes slice
		if len(step.StepRunnerConfig.Volumes) > 0 {
			cfg.Volumes = make([]string, len(step.StepRunnerConfig.Volumes))
			for i := range step.StepRunnerConfig.Volumes {
				cfg.Volumes[i] = get(fmt.Sprintf("RunnerConfig.Volumes[%d]", i))
			}
		}
		rendered.StepRunnerConfig = &core.StepRunnerConfig{RunnerConfig: &cfg}
	}

	// LLM messages handled separately (complex nested structure)
	if len(step.Messages) > 0 {
		if err := d.renderLLMMessages(&rendered, vars); err != nil {
			return nil, err
		}
	}

	return &rendered, nil
}

// renderLLMMessages renders LLM message templates (complex nested structure)
func (d *StepDispatcher) renderLLMMessages(rendered *core.Step, vars map[string]any) error {
	renderedMessages := make([]core.LLMMessage, len(rendered.Messages))
	for i, msg := range rendered.Messages {
		renderedMsg := msg

		// Render content (can be string or []interface{})
		switch content := msg.Content.(type) {
		case string:
			renderedContent, err := d.templateEngine.Render(content, vars)
			if err != nil {
				return fmt.Errorf("error rendering message content: %w", err)
			}
			renderedMsg.Content = renderedContent
		case []interface{}:
			// Handle multimodal content parts
			renderedParts := make([]interface{}, len(content))
			for j, part := range content {
				if partMap, ok := part.(map[string]interface{}); ok {
					renderedPartMap := make(map[string]interface{})
					for k, v := range partMap {
						renderedPartMap[k] = v
					}
					// Render text field
					if text, ok := partMap["text"].(string); ok {
						renderedText, err := d.templateEngine.Render(text, vars)
						if err != nil {
							return fmt.Errorf("error rendering content part text: %w", err)
						}
						renderedPartMap["text"] = renderedText
					}
					// Render image_url.url if present
					if imgURL, ok := partMap["image_url"].(map[string]interface{}); ok {
						renderedImgURL := make(map[string]interface{})
						for k, v := range imgURL {
							renderedImgURL[k] = v
						}
						if url, ok := imgURL["url"].(string); ok {
							renderedURL, err := d.templateEngine.Render(url, vars)
							if err != nil {
								return fmt.Errorf("error rendering image URL: %w", err)
							}
							renderedImgURL["url"] = renderedURL
						}
						renderedPartMap["image_url"] = renderedImgURL
					}
					renderedParts[j] = renderedPartMap
				} else {
					renderedParts[j] = part
				}
			}
			renderedMsg.Content = renderedParts
		}

		renderedMessages[i] = renderedMsg
	}
	rendered.Messages = renderedMessages
	return nil
}

// renderStepSequential renders step templates sequentially (original implementation)
func (d *StepDispatcher) renderStepSequential(step *core.Step, vars map[string]any) (*core.Step, error) {
	// Create a copy of the step
	rendered := *step

	// Render command fields
	if step.Command != "" {
		cmd, err := d.templateEngine.Render(step.Command, vars)
		if err != nil {
			return nil, err
		}
		rendered.Command = cmd
	}

	if len(step.Commands) > 0 {
		cmds, err := d.templateEngine.RenderSlice(step.Commands, vars)
		if err != nil {
			return nil, err
		}
		rendered.Commands = cmds
	}

	if len(step.ParallelCommands) > 0 {
		cmds, err := d.templateEngine.RenderSlice(step.ParallelCommands, vars)
		if err != nil {
			return nil, err
		}
		rendered.ParallelCommands = cmds
	}

	// Render structured argument fields (for bash/remote-bash steps)
	if step.SpeedArgs != "" {
		args, err := d.templateEngine.Render(step.SpeedArgs, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering speed_args: %w", err)
		}
		rendered.SpeedArgs = args
	}
	if step.ConfigArgs != "" {
		args, err := d.templateEngine.Render(step.ConfigArgs, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering config_args: %w", err)
		}
		rendered.ConfigArgs = args
	}
	if step.InputArgs != "" {
		args, err := d.templateEngine.Render(step.InputArgs, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering input_args: %w", err)
		}
		rendered.InputArgs = args
	}
	if step.OutputArgs != "" {
		args, err := d.templateEngine.Render(step.OutputArgs, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering output_args: %w", err)
		}
		rendered.OutputArgs = args
	}

	// Render std_file for stdout/stderr capture
	if step.StdFile != "" {
		stdFile, err := d.templateEngine.Render(step.StdFile, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering std_file: %w", err)
		}
		rendered.StdFile = stdFile
	}

	// Render function fields
	if step.Function != "" {
		fn, err := d.templateEngine.Render(step.Function, vars)
		if err != nil {
			return nil, err
		}
		rendered.Function = fn
	}

	if len(step.Functions) > 0 {
		fns, err := d.templateEngine.RenderSlice(step.Functions, vars)
		if err != nil {
			return nil, err
		}
		rendered.Functions = fns
	}

	if len(step.ParallelFunctions) > 0 {
		fns, err := d.templateEngine.RenderSlice(step.ParallelFunctions, vars)
		if err != nil {
			return nil, err
		}
		rendered.ParallelFunctions = fns
	}

	// Render foreach fields
	if step.Input != "" {
		input, err := d.templateEngine.Render(step.Input, vars)
		if err != nil {
			return nil, err
		}
		rendered.Input = input
	}

	// Render log message
	if step.Log != "" {
		log, err := d.templateEngine.Render(step.Log, vars)
		if err != nil {
			return nil, err
		}
		rendered.Log = log
	}

	if step.Timeout != "" {
		to, err := d.templateEngine.Render(string(step.Timeout), vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering timeout: %w", err)
		}
		rendered.Timeout = core.StepTimeout(to)
	}

	if step.Threads != "" {
		th, err := d.templateEngine.Render(string(step.Threads), vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering threads: %w", err)
		}
		rendered.Threads = core.StepThreads(th)
	}

	// Render HTTP step fields
	if step.URL != "" {
		url, err := d.templateEngine.Render(step.URL, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering url: %w", err)
		}
		rendered.URL = url
	}
	if step.Method != "" {
		method, err := d.templateEngine.Render(step.Method, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering method: %w", err)
		}
		rendered.Method = method
	}
	if step.RequestBody != "" {
		body, err := d.templateEngine.Render(step.RequestBody, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering request_body: %w", err)
		}
		rendered.RequestBody = body
	}
	if len(step.Headers) > 0 {
		headers, err := d.templateEngine.RenderMap(step.Headers, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering headers: %w", err)
		}
		rendered.Headers = headers
	}

	// Render step_runner if it contains template variables
	if step.StepRunner != "" {
		sr, err := d.templateEngine.Render(string(step.StepRunner), vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering step_runner: %w", err)
		}
		rendered.StepRunner = core.RunnerType(sr)
	}

	// Render step_runner_config fields for remote-bash steps
	if step.StepRunnerConfig != nil {
		renderedConfig := &core.StepRunnerConfig{}

		if step.StepRunnerConfig.RunnerConfig != nil {
			cfg := *step.StepRunnerConfig.RunnerConfig

			// Render string fields that may contain templates
			if cfg.Image != "" {
				img, err := d.templateEngine.Render(cfg.Image, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.image: %w", err)
				}
				cfg.Image = img
			}
			if cfg.Host != "" {
				host, err := d.templateEngine.Render(cfg.Host, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.host: %w", err)
				}
				cfg.Host = host
			}
			if cfg.User != "" {
				user, err := d.templateEngine.Render(cfg.User, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.user: %w", err)
				}
				cfg.User = user
			}
			if cfg.Password != "" {
				pass, err := d.templateEngine.Render(cfg.Password, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.password: %w", err)
				}
				cfg.Password = pass
			}
			if cfg.KeyFile != "" {
				keyFile, err := d.templateEngine.Render(cfg.KeyFile, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.key_file: %w", err)
				}
				cfg.KeyFile = keyFile
			}
			if cfg.WorkDir != "" {
				workDir, err := d.templateEngine.Render(cfg.WorkDir, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.workdir: %w", err)
				}
				cfg.WorkDir = workDir
			}
			if cfg.Network != "" {
				network, err := d.templateEngine.Render(cfg.Network, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.network: %w", err)
				}
				cfg.Network = network
			}

			// Render env map values
			if len(cfg.Env) > 0 {
				renderedEnv, err := d.templateEngine.RenderMap(cfg.Env, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.env: %w", err)
				}
				cfg.Env = renderedEnv
			}

			// Render volumes slice
			if len(cfg.Volumes) > 0 {
				renderedVols, err := d.templateEngine.RenderSlice(cfg.Volumes, vars)
				if err != nil {
					return nil, fmt.Errorf("error rendering step_runner_config.volumes: %w", err)
				}
				cfg.Volumes = renderedVols
			}

			renderedConfig.RunnerConfig = &cfg
		}

		rendered.StepRunnerConfig = renderedConfig
	}

	// Render remote-bash file copy fields
	if step.StepRemoteFile != "" {
		remoteFile, err := d.templateEngine.Render(step.StepRemoteFile, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering step_remote_file: %w", err)
		}
		rendered.StepRemoteFile = remoteFile
	}
	if step.HostOutputFile != "" {
		hostFile, err := d.templateEngine.Render(step.HostOutputFile, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering host_output_file: %w", err)
		}
		rendered.HostOutputFile = hostFile
	}

	// Render LLM step fields
	if len(step.Messages) > 0 {
		if err := d.renderLLMMessages(&rendered, vars); err != nil {
			return nil, err
		}
	}

	// Render embedding input
	if len(step.EmbeddingInput) > 0 {
		embInputs, err := d.templateEngine.RenderSlice(step.EmbeddingInput, vars)
		if err != nil {
			return nil, fmt.Errorf("error rendering embedding_input: %w", err)
		}
		rendered.EmbeddingInput = embInputs
	}

	return &rendered, nil
}

// GetFunctionRegistry returns the function registry
func (d *StepDispatcher) GetFunctionRegistry() *functions.Registry {
	return d.functionRegistry
}

// GetTemplateEngine returns the template engine
func (d *StepDispatcher) GetTemplateEngine() template.TemplateEngine {
	return d.templateEngine
}
