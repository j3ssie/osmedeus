package database

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// AgentSession persists agent execution sessions for querying, debugging, and resuming
type AgentSession struct {
	bun.BaseModel `bun:"table:agent_sessions,alias:as"`

	ID               int64     `bun:"id,pk,autoincrement"`
	RunID            int64     `bun:"run_id"`
	StepName         string    `bun:"step_name,notnull"`
	Query            string    `bun:"query"`
	PlanContent      string    `bun:"plan_content"`
	FinalContent     string    `bun:"final_content"`
	Iterations       int       `bun:"iterations"`
	TotalTokens      int       `bun:"total_tokens"`
	PromptTokens     int       `bun:"prompt_tokens"`
	CompletionTokens int       `bun:"completion_tokens"`
	ToolCallsJSON    string    `bun:"tool_calls_json"`
	ConversationJSON string    `bun:"conversation_json"`
	Status           string    `bun:"status,notnull"`
	DurationMs       int64     `bun:"duration_ms"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

// CreateAgentSession inserts a new agent session record
func CreateAgentSession(ctx context.Context, session *AgentSession) error {
	if db == nil {
		return nil // No database configured, skip persistence
	}
	_, err := db.NewInsert().Model(session).Exec(ctx)
	return err
}

// GetAgentSessionsByRun returns all agent sessions for a given run
func GetAgentSessionsByRun(ctx context.Context, runID int64) ([]AgentSession, error) {
	var sessions []AgentSession
	err := db.NewSelect().
		Model(&sessions).
		Where("run_id = ?", runID).
		Order("created_at ASC").
		Scan(ctx)
	return sessions, err
}

// GetAgentSessionsByStep returns all agent sessions for a given step name
func GetAgentSessionsByStep(ctx context.Context, stepName string) ([]AgentSession, error) {
	var sessions []AgentSession
	err := db.NewSelect().
		Model(&sessions).
		Where("step_name = ?", stepName).
		Order("created_at DESC").
		Scan(ctx)
	return sessions, err
}
