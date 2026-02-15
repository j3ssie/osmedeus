package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// AgentExecutor implements the agentic loop for type: agent steps.
// It queries the LLM, handles tool calls by executing them via the ToolExecutorRegistry,
// feeds results back, and repeats until the LLM responds without tool calls or
// the max_iterations limit is reached.
type AgentExecutor struct {
	templateEngine   template.TemplateEngine
	functionRegistry *functions.Registry
	config           *config.Config
	silent           bool
	currentDepth     int // 0 = top-level
	maxDepth         int // default 3
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor(engine template.TemplateEngine, funcRegistry *functions.Registry) *AgentExecutor {
	return &AgentExecutor{
		templateEngine:   engine,
		functionRegistry: funcRegistry,
	}
}

// Name returns the executor name
func (e *AgentExecutor) Name() string {
	return "agent"
}

// StepTypes returns the step types this executor handles
func (e *AgentExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeAgent}
}

// SetConfig sets the application config
func (e *AgentExecutor) SetConfig(cfg *config.Config) {
	e.config = cfg
}

// SetSilent enables or disables silent mode
func (e *AgentExecutor) SetSilent(s bool) {
	e.silent = s
}

// SetDepthContext sets the current nesting depth and maximum allowed depth.
func (e *AgentExecutor) SetDepthContext(depth, maxDepth int) {
	e.currentDepth = depth
	e.maxDepth = maxDepth
}

// agentState tracks the agent's runtime state during execution
type agentState struct {
	messages         []ChatMessage
	totalTokens      int
	promptTokens     int
	completionTokens int
	iteration        int
	toolResults      []map[string]interface{}
	finalContent     string
	planContent      string                   // populated by planning stage
	goalResults      []map[string]interface{} // results from each goal in multi-goal mode
	toolRegistry     *ToolExecutorRegistry    // pluggable tool dispatch
	tokenMu          sync.Mutex               // protects token fields for concurrent sub-agent merging
}

// MergeTokens safely adds child agent token counts into this state.
// Thread-safe for concurrent sub-agent spawning.
func (s *agentState) MergeTokens(total, prompt, completion int) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	s.totalTokens += total
	s.promptTokens += prompt
	s.completionTokens += completion
}

// Execute runs the agent loop
func (e *AgentExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	log := logger.Get()
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	// Validate config
	if e.config == nil {
		return e.fail(result, fmt.Errorf("agent executor config not set"))
	}

	// Validate required fields: need query OR queries (not both)
	if step.Query == "" && len(step.Queries) == 0 {
		return e.fail(result, fmt.Errorf("agent step '%s' requires 'query' or 'queries' field", step.Name))
	}
	if step.Query != "" && len(step.Queries) > 0 {
		return e.fail(result, fmt.Errorf("agent step '%s' cannot have both 'query' and 'queries'", step.Name))
	}
	if step.MaxIterations <= 0 {
		return e.fail(result, fmt.Errorf("agent step '%s' requires 'max_iterations' > 0", step.Name))
	}
	if len(step.AgentTools) == 0 {
		return e.fail(result, fmt.Errorf("agent step '%s' requires 'agent_tools' field", step.Name))
	}

	// Parse output schema if specified
	var outputSchema *core.LLMResponseFormat
	if step.OutputSchema != "" {
		var err error
		outputSchema, err = core.ParseOutputSchema(step.OutputSchema)
		if err != nil {
			return e.fail(result, fmt.Errorf("agent step '%s': %w", step.Name, err))
		}
	}

	// Resolve agent tools to OpenAI-compatible schemas (with spawn_agent if sub-agents present)
	tools, err := core.ResolveAgentToolsWithSubAgents(step.AgentTools, step.SubAgents)
	if err != nil {
		return e.fail(result, fmt.Errorf("agent step '%s': %w", step.Name, err))
	}

	// Get merged LLM config
	llmConfig := e.getMergedConfig(step)

	// Compute effective max depth
	effectiveMaxDepth := step.MaxAgentDepth
	if effectiveMaxDepth <= 0 {
		effectiveMaxDepth = core.DefaultMaxAgentDepth
	}
	// If spawned as child, inherit parent's max depth
	if e.maxDepth > 0 {
		effectiveMaxDepth = e.maxDepth
	}

	// Initialize state before building tool registry (needed as parentState arg)
	state := &agentState{}

	// Build ToolExecutorRegistry with sub-agent support if needed
	var toolRegistry *ToolExecutorRegistry
	if len(step.SubAgents) > 0 {
		toolRegistry = BuildToolRegistryWithSubAgents(
			step.AgentTools, e.functionRegistry,
			e.templateEngine, e.config, e.silent,
			e.currentDepth, effectiveMaxDepth,
			state, step.SubAgents,
		)
	} else {
		toolRegistry = BuildToolRegistry(step.AgentTools, e.functionRegistry)
	}
	state.toolRegistry = toolRegistry

	// Load resumed conversation if specified
	if step.Memory != nil && step.Memory.ResumePath != "" {
		if err := e.loadConversation(state, step.Memory.ResumePath); err != nil {
			log.Warn("Failed to load conversation for resume, starting fresh",
				zap.String("path", step.Memory.ResumePath),
				zap.Error(err),
			)
		}
	}

	// Determine queries to execute (Group 3: multi-goal)
	queries := []string{step.Query}
	if len(step.Queries) > 0 {
		queries = step.Queries
	}

	// Run planning stage if configured (Group 2)
	if step.PlanPrompt != "" {
		planContent, err := e.executePlanningStage(ctx, state, step, llmConfig)
		if err != nil {
			log.Warn("Planning stage failed, continuing without plan",
				zap.String("step", step.Name),
				zap.Error(err),
			)
		} else {
			state.planContent = planContent
		}
	}

	// Execute each query (single or multi-goal)
	for goalIdx, query := range queries {
		if query == "" {
			continue
		}

		// Build initial messages for this goal
		e.initMessages(state, query, step, llmConfig, goalIdx)

		log.Debug("Starting agent loop",
			zap.String("step", step.Name),
			zap.Int("goal", goalIdx+1),
			zap.Int("total_goals", len(queries)),
			zap.Int("max_iterations", step.MaxIterations),
			zap.Int("tools", len(tools)),
		)

		// Main agent loop
		for state.iteration = 1; state.iteration <= step.MaxIterations; state.iteration++ {
			log.Debug("Agent iteration",
				zap.String("step", step.Name),
				zap.Int("iteration", state.iteration),
				zap.Int("messages", len(state.messages)),
			)

			// Build LLM request, potentially with structured output on final iteration
			var responseFormat *core.LLMResponseFormat
			if outputSchema != nil && state.iteration == step.MaxIterations {
				responseFormat = outputSchema
			}

			// Call LLM with optional model fallback (Group 5)
			response, err := e.callLLMWithFallback(ctx, state, tools, llmConfig, step.Models, responseFormat)
			if err != nil {
				return e.fail(result, fmt.Errorf("agent step '%s' iteration %d: %w", step.Name, state.iteration, err))
			}

			if response == nil || len(response.Choices) == 0 {
				return e.fail(result, fmt.Errorf("agent step '%s': empty response from LLM", step.Name))
			}

			// Track tokens
			state.totalTokens += response.Usage.TotalTokens
			state.promptTokens += response.Usage.PromptTokens
			state.completionTokens += response.Usage.CompletionTokens

			choice := response.Choices[0]

			// Append assistant message to conversation
			state.messages = append(state.messages, choice.Message)

			// Extract content
			if content, ok := choice.Message.Content.(string); ok {
				state.finalContent = content
			}

			// Check if we're done (no tool calls)
			if len(choice.Message.ToolCalls) == 0 {
				// If we have OutputSchema and this isn't the max iteration,
				// try to enforce structured output on the next call
				if outputSchema != nil && state.iteration < step.MaxIterations {
					log.Debug("Agent completed (no tool calls), structured output available",
						zap.String("step", step.Name),
						zap.Int("iterations", state.iteration),
					)
				} else {
					log.Debug("Agent completed (no tool calls)",
						zap.String("step", step.Name),
						zap.Int("iterations", state.iteration),
					)
				}
				break
			}

			// Execute tool calls with tracing hooks (Group 7)
			toolMessages, err := e.executeToolCalls(ctx, choice.Message.ToolCalls, state, step, execCtx)
			if err != nil {
				return e.fail(result, fmt.Errorf("agent step '%s' iteration %d tool execution: %w", step.Name, state.iteration, err))
			}

			// Append tool results to conversation
			state.messages = append(state.messages, toolMessages...)

			// Evaluate stop condition if defined
			if step.StopCondition != "" {
				vars := execCtx.GetVariables()
				vars["agent_content"] = state.finalContent
				vars["iteration"] = state.iteration
				shouldStop, err := e.functionRegistry.EvaluateCondition(step.StopCondition, vars)
				if err != nil {
					log.Warn("Stop condition evaluation failed",
						zap.String("step", step.Name),
						zap.Error(err),
					)
				} else if shouldStop {
					log.Debug("Agent stopped by stop_condition",
						zap.String("step", step.Name),
						zap.Int("iteration", state.iteration),
					)
					break
				}
			}

			// Apply sliding window if configured (with optional compression — Group 4)
			if step.Memory != nil && step.Memory.MaxMessages > 0 {
				if step.Memory.SummarizeOnTruncate {
					e.applyMessageWindowWithSummary(ctx, state, step.Memory.MaxMessages, llmConfig)
				} else {
					e.applyMessageWindow(state, step.Memory.MaxMessages)
				}
			}
		}

		// Record goal result (Group 3)
		goalResult := map[string]interface{}{
			"query":   query,
			"content": state.finalContent,
		}
		state.goalResults = append(state.goalResults, goalResult)
	}

	// If OutputSchema is set and we haven't gotten structured output yet,
	// make a final structured output request (Group 6)
	if outputSchema != nil {
		structuredContent := e.requestStructuredOutput(ctx, state, outputSchema, llmConfig)
		if structuredContent != "" {
			state.finalContent = structuredContent
		}
	}

	// Print final output (skip if streaming — tokens were already printed in real-time)
	if !e.silent && !llmConfig.Stream && state.finalContent != "" {
		printLLMOutput(state.finalContent)
	}

	// Persist conversation if configured
	if step.Memory != nil && step.Memory.PersistPath != "" {
		if err := e.persistConversation(state, step.Memory.PersistPath); err != nil {
			log.Warn("Failed to persist agent conversation",
				zap.String("path", step.Memory.PersistPath),
				zap.Error(err),
			)
		}
	}

	// Set exports
	historyJSON, err := json.Marshal(state.messages)
	if err != nil {
		log.Warn("Failed to marshal agent history", zap.Error(err))
		historyJSON = []byte("[]")
	}
	toolResultsJSON, err := json.Marshal(state.toolResults)
	if err != nil {
		log.Warn("Failed to marshal agent tool results", zap.Error(err))
		toolResultsJSON = []byte("[]")
	}

	result.Exports["agent_content"] = state.finalContent
	result.Exports["agent_history"] = string(historyJSON)
	// Cap iteration count: the for-loop post-increments past MaxIterations on natural exit
	iterations := state.iteration
	if iterations > step.MaxIterations {
		iterations = step.MaxIterations
	}
	result.Exports["agent_iterations"] = iterations
	result.Exports["agent_total_tokens"] = state.totalTokens
	result.Exports["agent_prompt_tokens"] = state.promptTokens
	result.Exports["agent_completion_tokens"] = state.completionTokens
	result.Exports["agent_tool_results"] = string(toolResultsJSON)

	// Planning stage export (Group 2)
	if state.planContent != "" {
		result.Exports["agent_plan"] = state.planContent
	}

	// Multi-goal results export (Group 3)
	if len(state.goalResults) > 1 {
		goalResultsJSON, err := json.Marshal(state.goalResults)
		if err != nil {
			log.Warn("Failed to marshal goal results", zap.Error(err))
			goalResultsJSON = []byte("[]")
		}
		result.Exports["agent_goal_results"] = string(goalResultsJSON)
	}

	result.Output = state.finalContent
	result.Status = core.StepStatusSuccess
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// executePlanningStage runs the planning phase before the main agent loop (Group 2)
func (e *AgentExecutor) executePlanningStage(
	ctx context.Context,
	state *agentState,
	step *core.Step,
	llmConfig *MergedLLMConfig,
) (string, error) {
	log := logger.Get()
	log.Debug("Executing planning stage",
		zap.String("step", step.Name),
	)

	// Build planning messages
	var planMessages []ChatMessage

	// System prompt if available
	systemPrompt := step.SystemPrompt
	if systemPrompt == "" && llmConfig.SystemPrompt != "" {
		systemPrompt = llmConfig.SystemPrompt
	}
	if systemPrompt != "" {
		planMessages = append(planMessages, ChatMessage{
			Role:    string(core.LLMRoleSystem),
			Content: systemPrompt,
		})
	}

	// Plan prompt as user message
	planMessages = append(planMessages, ChatMessage{
		Role:    string(core.LLMRoleUser),
		Content: step.PlanPrompt,
	})

	// Build planning request (no tools — just text generation)
	planConfig := *llmConfig
	if step.PlanMaxTokens != nil {
		planConfig.MaxTokens = *step.PlanMaxTokens
	}

	planState := &agentState{messages: planMessages}
	response, err := e.callLLM(ctx, planState, nil, &planConfig)
	if err != nil {
		return "", fmt.Errorf("planning request failed: %w", err)
	}

	if response == nil || len(response.Choices) == 0 {
		return "", fmt.Errorf("empty response from planning request")
	}

	// Track tokens from planning
	state.totalTokens += response.Usage.TotalTokens
	state.promptTokens += response.Usage.PromptTokens
	state.completionTokens += response.Usage.CompletionTokens

	planContent := ""
	if content, ok := response.Choices[0].Message.Content.(string); ok {
		planContent = content
	}

	log.Debug("Planning stage complete",
		zap.String("step", step.Name),
		zap.Int("plan_length", len(planContent)),
	)

	return planContent, nil
}

// initMessages builds the initial conversation messages for a goal
func (e *AgentExecutor) initMessages(state *agentState, query string, step *core.Step, llmConfig *MergedLLMConfig, goalIdx int) {
	// For the first goal, build from scratch or resume
	if goalIdx == 0 {
		// If we loaded resumed messages, just append the new query
		if len(state.messages) > 0 {
			state.messages = append(state.messages, ChatMessage{
				Role:    string(core.LLMRoleUser),
				Content: query,
			})
			return
		}

		// Start fresh conversation
		systemPrompt := step.SystemPrompt
		if systemPrompt == "" && llmConfig.SystemPrompt != "" {
			systemPrompt = llmConfig.SystemPrompt
		}
		if systemPrompt != "" {
			state.messages = append(state.messages, ChatMessage{
				Role:    string(core.LLMRoleSystem),
				Content: systemPrompt,
			})
		}

		// Prepend plan if available (Group 2)
		if state.planContent != "" {
			state.messages = append(state.messages, ChatMessage{
				Role:    string(core.LLMRoleAssistant),
				Content: "Here is my plan:\n\n" + state.planContent,
			})
		}

		// User query
		state.messages = append(state.messages, ChatMessage{
			Role:    string(core.LLMRoleUser),
			Content: query,
		})
		return
	}

	// For subsequent goals in multi-goal mode, append new user message
	state.messages = append(state.messages, ChatMessage{
		Role:    string(core.LLMRoleUser),
		Content: query,
	})
}

// callLLMWithFallback calls the LLM with optional per-agent model fallback (Group 5)
func (e *AgentExecutor) callLLMWithFallback(
	ctx context.Context,
	state *agentState,
	tools []core.LLMTool,
	llmConfig *MergedLLMConfig,
	models []string,
	responseFormat *core.LLMResponseFormat,
) (*ChatCompletionResponse, error) {
	// If step specifies preferred models, try each in order
	if len(models) > 0 {
		var lastErr error
		for _, model := range models {
			modelConfig := *llmConfig
			modelConfig.Model = model
			if responseFormat != nil {
				modelConfig.ResponseFormat = responseFormat
			}
			response, err := e.callLLM(ctx, state, tools, &modelConfig)
			if err == nil && (response == nil || response.Error == nil) {
				return response, nil
			}
			lastErr = err
			logger.Get().Debug("Model fallback: trying next model",
				zap.String("failed_model", model),
				zap.Error(err),
			)
		}
		// Fall through to default provider rotation
		logger.Get().Warn("All specified models failed, falling back to default",
			zap.Error(lastErr),
		)
	}

	// Default path: use standard provider rotation
	if responseFormat != nil {
		configWithFormat := *llmConfig
		configWithFormat.ResponseFormat = responseFormat
		return e.callLLM(ctx, state, tools, &configWithFormat)
	}
	return e.callLLM(ctx, state, tools, llmConfig)
}

// callLLM sends the current conversation to the LLM and returns the response
func (e *AgentExecutor) callLLM(
	ctx context.Context,
	state *agentState,
	tools []core.LLMTool,
	llmConfig *MergedLLMConfig,
) (*ChatCompletionResponse, error) {
	log := logger.Get()

	request := &ChatCompletionRequest{
		Model:          llmConfig.Model,
		Messages:       state.messages,
		MaxTokens:      llmConfig.MaxTokens,
		Temperature:    llmConfig.Temperature,
		TopP:           llmConfig.TopP,
		TopK:           llmConfig.TopK,
		Tools:          tools,
		Stream:         llmConfig.Stream,
		ResponseFormat: llmConfig.ResponseFormat,
	}

	// Provider rotation with retries
	var response *ChatCompletionResponse
	var lastErr error

	maxRetries := llmConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	providerCount := e.config.LLM.GetProviderCount()
	if providerCount == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}

	totalAttempts := maxRetries * providerCount

	for attempt := 0; attempt < totalAttempts; attempt++ {
		provider := e.config.LLM.GetCurrentProvider()
		if provider == nil {
			lastErr = fmt.Errorf("no LLM providers available")
			break
		}

		if llmConfig.Model == "" {
			request.Model = provider.Model
		}

		log.Debug("Agent LLM request",
			zap.String("provider", provider.Provider),
			zap.String("model", request.Model),
			zap.Int("attempt", attempt+1),
		)

		response, lastErr = e.sendChatRequest(ctx, provider, request, llmConfig)

		if lastErr == nil && response.Error == nil {
			break
		}

		if isProviderError(lastErr) || isRateLimitError(response) {
			e.config.LLM.RotateProvider()
		}

		if attempt < totalAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 500 * time.Millisecond):
			}
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	if response != nil && response.Error != nil {
		return nil, fmt.Errorf("LLM API error: %s (%s)", response.Error.Message, response.Error.Type)
	}

	return response, nil
}

// sendChatRequest sends an HTTP request to the LLM provider (delegates to LLMExecutor's logic)
func (e *AgentExecutor) sendChatRequest(
	ctx context.Context,
	provider *config.LLMProvider,
	request *ChatCompletionRequest,
	llmConfig *MergedLLMConfig,
) (*ChatCompletionResponse, error) {
	llmExec := &LLMExecutor{config: e.config}
	return llmExec.sendChatRequest(ctx, provider, request, llmConfig)
}

// executeToolCalls executes tool calls from the LLM response, with tracing hooks (Group 7)
func (e *AgentExecutor) executeToolCalls(
	ctx context.Context,
	toolCalls []core.LLMToolCall,
	state *agentState,
	step *core.Step,
	execCtx *core.ExecutionContext,
) ([]ChatMessage, error) {
	log := logger.Get()

	parallelToolCalls := true
	if step.ParallelToolCalls != nil {
		parallelToolCalls = *step.ParallelToolCalls
	}

	if parallelToolCalls && len(toolCalls) > 1 {
		return e.executeToolCallsParallel(ctx, toolCalls, state, step, execCtx, log)
	}

	return e.executeToolCallsSequential(ctx, toolCalls, state, step, execCtx, log)
}

// executeToolCallsSequential executes tool calls one at a time
func (e *AgentExecutor) executeToolCallsSequential(
	ctx context.Context,
	toolCalls []core.LLMToolCall,
	state *agentState,
	step *core.Step,
	execCtx *core.ExecutionContext,
	log *zap.Logger,
) ([]ChatMessage, error) {
	messages := make([]ChatMessage, 0, len(toolCalls))

	for _, tc := range toolCalls {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute on_tool_start hook (Group 7)
		e.executeToolHook(step.OnToolStart, tc.Function.Name, tc.Function.Arguments, "", 0, state.iteration, nil, execCtx)

		startTime := time.Now()
		result, err := e.executeSingleToolCall(ctx, tc, state, execCtx, log)
		duration := time.Since(startTime)
		if err != nil {
			result = fmt.Sprintf("Error executing tool '%s': %s", tc.Function.Name, err.Error())
		}

		// Execute on_tool_end hook (Group 7)
		e.executeToolHook(step.OnToolEnd, tc.Function.Name, tc.Function.Arguments, result, duration.Milliseconds(), state.iteration, err, execCtx)

		log.Debug("Tool call completed",
			zap.String("tool", tc.Function.Name),
			zap.String("call_id", tc.ID),
			zap.Duration("duration", duration),
		)

		state.toolResults = append(state.toolResults, map[string]interface{}{
			"tool_call_id": tc.ID,
			"tool_name":    tc.Function.Name,
			"result":       result,
		})

		messages = append(messages, ChatMessage{
			Role:       string(core.LLMRoleTool),
			Content:    result,
			ToolCallID: tc.ID,
		})
	}

	return messages, nil
}

// executeToolCallsParallel executes tool calls concurrently
func (e *AgentExecutor) executeToolCallsParallel(
	ctx context.Context,
	toolCalls []core.LLMToolCall,
	state *agentState,
	step *core.Step,
	execCtx *core.ExecutionContext,
	log *zap.Logger,
) ([]ChatMessage, error) {
	type toolResult struct {
		index   int
		message ChatMessage
	}

	results := make(chan toolResult, len(toolCalls))
	var wg sync.WaitGroup

	for i, tc := range toolCalls {
		wg.Add(1)
		go func(idx int, call core.LLMToolCall) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			// Execute on_tool_start hook (Group 7)
			e.executeToolHook(step.OnToolStart, call.Function.Name, call.Function.Arguments, "", 0, state.iteration, nil, execCtx)

			startTime := time.Now()
			result, err := e.executeSingleToolCall(ctx, call, state, execCtx, log)
			duration := time.Since(startTime)
			if err != nil {
				result = fmt.Sprintf("Error executing tool '%s': %s", call.Function.Name, err.Error())
			}

			// Execute on_tool_end hook (Group 7)
			e.executeToolHook(step.OnToolEnd, call.Function.Name, call.Function.Arguments, result, duration.Milliseconds(), state.iteration, err, execCtx)

			log.Debug("Tool call completed",
				zap.String("tool", call.Function.Name),
				zap.String("call_id", call.ID),
				zap.Duration("duration", duration),
			)

			results <- toolResult{
				index: idx,
				message: ChatMessage{
					Role:       string(core.LLMRoleTool),
					Content:    result,
					ToolCallID: call.ID,
				},
			}
		}(i, tc)
	}

	wg.Wait()
	close(results)

	// Collect and order results, populate toolResults
	ordered := make([]ChatMessage, len(toolCalls))
	for r := range results {
		ordered[r.index] = r.message
	}
	for i, msg := range ordered {
		state.toolResults = append(state.toolResults, map[string]interface{}{
			"tool_call_id": toolCalls[i].ID,
			"tool_name":    toolCalls[i].Function.Name,
			"result":       msg.Content,
		})
	}

	return ordered, nil
}

// executeSingleToolCall executes a single tool call via the ToolExecutorRegistry
func (e *AgentExecutor) executeSingleToolCall(
	ctx context.Context,
	tc core.LLMToolCall,
	state *agentState,
	execCtx *core.ExecutionContext,
	log *zap.Logger,
) (string, error) {
	// Use ToolExecutorRegistry if available (Group 1)
	if state.toolRegistry != nil {
		return executeToolCallViaRegistry(ctx, tc, state.toolRegistry, execCtx, log)
	}

	// Fallback to legacy dispatch (should not happen in normal flow)
	return e.executeSingleToolCallLegacy(tc, execCtx, log)
}

// executeSingleToolCallLegacy is the legacy dispatch path (kept for safety)
func (e *AgentExecutor) executeSingleToolCallLegacy(
	tc core.LLMToolCall,
	execCtx *core.ExecutionContext,
	log *zap.Logger,
) (string, error) {
	funcName := tc.Function.Name
	argsJSON := tc.Function.Arguments

	log.Debug("Executing tool call (legacy)",
		zap.String("tool", funcName),
		zap.String("args", argsJSON),
	)

	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments for %s: %w", funcName, err)
		}
	}
	if args == nil {
		args = make(map[string]interface{})
	}

	return e.executePresetTool(funcName, args, execCtx)
}

// executePresetTool runs a preset tool by building a function call expression
func (e *AgentExecutor) executePresetTool(funcName string, args map[string]interface{}, execCtx *core.ExecutionContext) (string, error) {
	expr := buildPresetCallExpr(funcName, args)
	vars := execCtx.GetVariables()
	result, err := e.functionRegistry.Execute(expr, vars)
	if err != nil {
		return "", fmt.Errorf("preset tool '%s' failed: %w", funcName, err)
	}
	return formatToolResult(result), nil
}

// executeToolHook evaluates a JS hook expression with tool call context (Group 7)
func (e *AgentExecutor) executeToolHook(
	hook string,
	toolName string,
	toolArgs string,
	result string,
	durationMs int64,
	iteration int,
	toolErr error,
	execCtx *core.ExecutionContext,
) {
	if hook == "" || e.functionRegistry == nil {
		return
	}

	log := logger.Get()
	vars := execCtx.GetVariables()
	vars["tool_name"] = toolName
	vars["tool_args"] = toolArgs
	vars["result"] = result
	vars["duration"] = durationMs
	vars["iteration"] = iteration
	if toolErr != nil {
		vars["error"] = toolErr.Error()
	} else {
		vars["error"] = ""
	}

	if _, err := e.functionRegistry.Execute(hook, vars); err != nil {
		log.Warn("Tool hook execution failed (non-blocking)",
			zap.String("hook", hook),
			zap.Error(err),
		)
	}
}

// requestStructuredOutput makes a final LLM call to enforce structured output (Group 6)
func (e *AgentExecutor) requestStructuredOutput(
	ctx context.Context,
	state *agentState,
	schema *core.LLMResponseFormat,
	llmConfig *MergedLLMConfig,
) string {
	log := logger.Get()

	// Only request if we have a schema and the last response might not be structured
	if schema == nil {
		return ""
	}

	// Check if current content is already valid JSON matching the schema
	var check interface{}
	if json.Unmarshal([]byte(state.finalContent), &check) == nil {
		// Already valid JSON, likely structured
		return ""
	}

	// Add instruction to produce structured output
	state.messages = append(state.messages, ChatMessage{
		Role:    string(core.LLMRoleUser),
		Content: "Please provide your final answer in the structured JSON format specified.",
	})

	configWithSchema := *llmConfig
	configWithSchema.ResponseFormat = schema

	response, err := e.callLLM(ctx, state, nil, &configWithSchema)
	if err != nil {
		log.Warn("Structured output request failed",
			zap.Error(err),
		)
		return ""
	}

	if response == nil || len(response.Choices) == 0 {
		return ""
	}

	// Track tokens
	state.totalTokens += response.Usage.TotalTokens
	state.promptTokens += response.Usage.PromptTokens
	state.completionTokens += response.Usage.CompletionTokens

	if content, ok := response.Choices[0].Message.Content.(string); ok {
		return content
	}
	return ""
}

// buildPresetCallExpr builds a JS function call expression from a preset tool name and arguments
func buildPresetCallExpr(funcName string, args map[string]interface{}) string {
	switch funcName {
	case "bash":
		return fmt.Sprintf("bash(%s)", jsQuote(getStringArg(args, "command")))
	case "read_file":
		return fmt.Sprintf("read_file(%s)", jsQuote(getStringArg(args, "path")))
	case "read_lines":
		return fmt.Sprintf("read_lines(%s)", jsQuote(getStringArg(args, "path")))
	case "file_exists":
		return fmt.Sprintf("file_exists(%s)", jsQuote(getStringArg(args, "path")))
	case "file_length":
		return fmt.Sprintf("file_length(%s)", jsQuote(getStringArg(args, "path")))
	case "append_file":
		return fmt.Sprintf("append_file(%s, %s)", jsQuote(getStringArg(args, "dest")), jsQuote(getStringArg(args, "content")))
	case "save_content":
		return fmt.Sprintf("save_content(%s, %s)", jsQuote(getStringArg(args, "content")), jsQuote(getStringArg(args, "path")))
	case "glob":
		return fmt.Sprintf("glob(%s)", jsQuote(getStringArg(args, "pattern")))
	case "grep_string":
		return fmt.Sprintf("grep_string(%s, %s)", jsQuote(getStringArg(args, "source")), jsQuote(getStringArg(args, "str")))
	case "grep_regex":
		return fmt.Sprintf("grep_regex(%s, %s)", jsQuote(getStringArg(args, "source")), jsQuote(getStringArg(args, "pattern")))
	case "http_get":
		return fmt.Sprintf("http_get(%s)", jsQuote(getStringArg(args, "url")))
	case "http_request":
		return fmt.Sprintf("http_request(%s, %s, %s, %s)",
			jsQuote(getStringArg(args, "url")),
			jsQuote(getStringArg(args, "method")),
			jsQuote(getStringArg(args, "headers")),
			jsQuote(getStringArg(args, "body")),
		)
	case "jq":
		return fmt.Sprintf("jq(%s, %s)", jsQuote(getStringArg(args, "json_data")), jsQuote(getStringArg(args, "expression")))
	case "exec_python":
		return fmt.Sprintf("exec_python(%s)", jsQuote(getStringArg(args, "code")))
	case "exec_python_file":
		return fmt.Sprintf("exec_python_file(%s)", jsQuote(getStringArg(args, "path")))
	case "exec_ts":
		return fmt.Sprintf("exec_ts(%s)", jsQuote(getStringArg(args, "code")))
	case "exec_ts_file":
		return fmt.Sprintf("exec_ts_file(%s)", jsQuote(getStringArg(args, "path")))
	case "run_module":
		return fmt.Sprintf("run_module(%s, %s, %s)", jsQuote(getStringArg(args, "module")), jsQuote(getStringArg(args, "target")), jsQuote(getStringArg(args, "params")))
	case "run_flow":
		return fmt.Sprintf("run_flow(%s, %s, %s)", jsQuote(getStringArg(args, "flow")), jsQuote(getStringArg(args, "target")), jsQuote(getStringArg(args, "params")))
	default:
		var argStrs []string
		for _, v := range args {
			argStrs = append(argStrs, jsQuote(fmt.Sprintf("%v", v)))
		}
		return fmt.Sprintf("%s(%s)", funcName, strings.Join(argStrs, ", "))
	}
}

// getStringArg safely extracts a string argument from the args map
func getStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// jsQuote returns a JavaScript string literal, escaping special characters
func jsQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return `"` + s + `"`
}

// formatToolResult converts a tool execution result to a string for the LLM
func formatToolResult(result interface{}) string {
	if result == nil {
		return ""
	}

	switch v := result.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	}
}

// applyMessageWindow trims the conversation to max_messages, keeping the system message
func (e *AgentExecutor) applyMessageWindow(state *agentState, maxMessages int) {
	if maxMessages <= 0 || len(state.messages) <= maxMessages {
		return
	}

	var systemMsg *ChatMessage
	startIdx := 0
	if len(state.messages) > 0 && state.messages[0].Role == string(core.LLMRoleSystem) {
		systemMsg = &state.messages[0]
		startIdx = 1
	}

	nonSystemMsgs := state.messages[startIdx:]
	keepCount := maxMessages
	if systemMsg != nil {
		keepCount--
	}

	if len(nonSystemMsgs) > keepCount {
		nonSystemMsgs = nonSystemMsgs[len(nonSystemMsgs)-keepCount:]
	}

	if systemMsg != nil {
		state.messages = make([]ChatMessage, 0, keepCount+1)
		state.messages = append(state.messages, *systemMsg)
		state.messages = append(state.messages, nonSystemMsgs...)
	} else {
		state.messages = nonSystemMsgs
	}
}

// applyMessageWindowWithSummary trims with LLM-based summarization of dropped messages (Group 4)
func (e *AgentExecutor) applyMessageWindowWithSummary(
	ctx context.Context,
	state *agentState,
	maxMessages int,
	llmConfig *MergedLLMConfig,
) {
	if maxMessages <= 0 || len(state.messages) <= maxMessages {
		return
	}

	log := logger.Get()

	var systemMsg *ChatMessage
	startIdx := 0
	if len(state.messages) > 0 && state.messages[0].Role == string(core.LLMRoleSystem) {
		systemMsg = &state.messages[0]
		startIdx = 1
	}

	nonSystemMsgs := state.messages[startIdx:]
	keepCount := maxMessages
	if systemMsg != nil {
		keepCount-- // Account for system message
		keepCount-- // Account for summary message we'll insert
	}
	if keepCount < 1 {
		keepCount = 1
	}

	if len(nonSystemMsgs) <= keepCount {
		return
	}

	// Messages to be dropped
	dropCount := len(nonSystemMsgs) - keepCount
	droppedMsgs := nonSystemMsgs[:dropCount]
	keptMsgs := nonSystemMsgs[dropCount:]

	// Build summary of dropped messages
	var summaryParts []string
	for _, msg := range droppedMsgs {
		if content, ok := msg.Content.(string); ok && content != "" {
			role := msg.Role
			// Truncate long messages
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			summaryParts = append(summaryParts, fmt.Sprintf("[%s]: %s", role, content))
		}
	}

	if len(summaryParts) == 0 {
		// No meaningful content to summarize, just truncate
		e.applyMessageWindow(state, maxMessages)
		return
	}

	// Ask LLM to summarize the dropped context
	summaryPrompt := "Summarize the following conversation context concisely, preserving key information:\n\n" + strings.Join(summaryParts, "\n")
	summaryMessages := []ChatMessage{
		{Role: string(core.LLMRoleUser), Content: summaryPrompt},
	}

	summaryState := &agentState{messages: summaryMessages}
	summaryConfig := *llmConfig
	summaryConfig.MaxTokens = 300

	response, err := e.callLLM(ctx, summaryState, nil, &summaryConfig)
	if err != nil {
		log.Warn("Conversation summarization failed, falling back to simple truncation",
			zap.Error(err),
		)
		e.applyMessageWindow(state, maxMessages)
		return
	}

	summaryContent := ""
	if response != nil && len(response.Choices) > 0 {
		if content, ok := response.Choices[0].Message.Content.(string); ok {
			summaryContent = content
		}
		// Track summarization tokens
		state.totalTokens += response.Usage.TotalTokens
		state.promptTokens += response.Usage.PromptTokens
		state.completionTokens += response.Usage.CompletionTokens
	}

	if summaryContent == "" {
		e.applyMessageWindow(state, maxMessages)
		return
	}

	// Rebuild messages with summary
	state.messages = make([]ChatMessage, 0, keepCount+2)
	if systemMsg != nil {
		state.messages = append(state.messages, *systemMsg)
	}
	state.messages = append(state.messages, ChatMessage{
		Role:    string(core.LLMRoleSystem),
		Content: "[Summary of earlier conversation]\n" + summaryContent,
	})
	state.messages = append(state.messages, keptMsgs...)
}

// loadConversation loads a prior conversation from a JSON file
func (e *AgentExecutor) loadConversation(state *agentState, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read conversation file: %w", err)
	}

	var messages []ChatMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to parse conversation file: %w", err)
	}

	state.messages = messages
	return nil
}

// persistConversation saves the conversation to a JSON file
func (e *AgentExecutor) persistConversation(state *agentState, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(state.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write conversation file: %w", err)
	}

	return nil
}

// getMergedConfig merges global and step-level LLM configuration
func (e *AgentExecutor) getMergedConfig(step *core.Step) *MergedLLMConfig {
	llmExec := &LLMExecutor{config: e.config}
	return llmExec.getMergedConfig(step)
}

// fail is a helper that sets the result to failed state
func (e *AgentExecutor) fail(result *core.StepResult, err error) (*core.StepResult, error) {
	result.Status = core.StepStatusFailed
	result.Error = err
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, err
}
