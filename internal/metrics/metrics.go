package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WorkflowDuration tracks the duration of workflow execution in seconds
	WorkflowDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "osmedeus_workflow_duration_seconds",
		Help:    "Duration of workflow execution in seconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 15), // 1s to ~9h
	}, []string{"workflow", "kind", "status"})

	// WorkflowsTotal counts the total number of workflows executed
	WorkflowsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "osmedeus_workflows_total",
		Help: "Total number of workflows executed",
	}, []string{"workflow", "kind", "status"})

	// ActiveRuns tracks the number of currently running workflows
	ActiveRuns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "osmedeus_active_runs",
		Help: "Number of currently running workflows",
	})

	// StepDuration tracks the duration of step execution in seconds
	StepDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "osmedeus_step_duration_seconds",
		Help:    "Duration of step execution in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 12), // 0.1s to ~7min
	}, []string{"step_type", "status"})

	// StepFailures counts the total number of step failures
	StepFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "osmedeus_step_failures_total",
		Help: "Total number of step failures",
	}, []string{"step_name", "step_type", "error_type"})

	// StepsTotal counts the total number of steps executed
	StepsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "osmedeus_steps_total",
		Help: "Total number of steps executed",
	}, []string{"step_type", "status"})

	// ToolExecutionDuration tracks external tool execution time
	ToolExecutionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "osmedeus_tool_execution_duration_seconds",
		Help:    "Duration of external tool execution in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15), // 0.1s to ~54min
	}, []string{"tool", "status"})

	// MemoryUsageBytes tracks current memory usage
	MemoryUsageBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "osmedeus_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	}, []string{"type"})

	// RateLimitHits counts rate limit encounters
	RateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "osmedeus_rate_limit_hits_total",
		Help: "Total number of rate limit hits encountered",
	}, []string{"target", "tool"})
)

// RecordWorkflowStart increments the active runs counter
func RecordWorkflowStart() {
	ActiveRuns.Inc()
}

// RecordWorkflowEnd decrements active runs and records duration/status
func RecordWorkflowEnd(workflowName, kind, status string, durationSeconds float64) {
	ActiveRuns.Dec()
	WorkflowDuration.WithLabelValues(workflowName, kind, status).Observe(durationSeconds)
	WorkflowsTotal.WithLabelValues(workflowName, kind, status).Inc()
}

// RecordStepDuration records the duration and status of a step execution
func RecordStepDuration(stepType, status string, durationSeconds float64) {
	StepDuration.WithLabelValues(stepType, status).Observe(durationSeconds)
	StepsTotal.WithLabelValues(stepType, status).Inc()
}

// RecordStepFailure records a step failure with error classification
func RecordStepFailure(stepName, stepType, errorType string) {
	StepFailures.WithLabelValues(stepName, stepType, errorType).Inc()
}

// RecordToolExecution records the duration and status of an external tool execution
func RecordToolExecution(toolName, status string, durationSeconds float64) {
	ToolExecutionDuration.WithLabelValues(toolName, status).Observe(durationSeconds)
}

// RecordRateLimitHit increments the rate limit hit counter
func RecordRateLimitHit(target, tool string) {
	RateLimitHits.WithLabelValues(target, tool).Inc()
}

// UpdateMemoryMetrics updates memory usage metrics from runtime.MemStats
func UpdateMemoryMetrics(m *runtime.MemStats) {
	MemoryUsageBytes.WithLabelValues("alloc").Set(float64(m.Alloc))
	MemoryUsageBytes.WithLabelValues("total_alloc").Set(float64(m.TotalAlloc))
	MemoryUsageBytes.WithLabelValues("sys").Set(float64(m.Sys))
	MemoryUsageBytes.WithLabelValues("heap_alloc").Set(float64(m.HeapAlloc))
	MemoryUsageBytes.WithLabelValues("heap_inuse").Set(float64(m.HeapInuse))
	MemoryUsageBytes.WithLabelValues("stack_inuse").Set(float64(m.StackInuse))
}
