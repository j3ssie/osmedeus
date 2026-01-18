package repository

import (
	"context"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/uptrace/bun"
)

// EventLogRepository handles event log database operations
type EventLogRepository struct {
	db *bun.DB
}

// NewEventLogRepository creates a new event log repository
func NewEventLogRepository(db *bun.DB) *EventLogRepository {
	return &EventLogRepository{db: db}
}

// EventLogQuery represents query parameters for event search
type EventLogQuery struct {
	Topic        string
	Name         string
	Source       string
	Workspace    string
	ScanID       string
	WorkflowName string
	Processed    *bool
	StartTime    *time.Time
	EndTime      *time.Time
	Page         int
	PerPage      int
}

// Create creates a new event log
func (r *EventLogRepository) Create(ctx context.Context, event *database.EventLog) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	_, err := r.db.NewInsert().Model(event).Exec(ctx)
	return err
}

// GetByID retrieves an event log by ID
func (r *EventLogRepository) GetByID(ctx context.Context, id int64) (*database.EventLog, error) {
	event := new(database.EventLog)
	err := r.db.NewSelect().
		Model(event).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// GetByEventID retrieves an event log by event ID (UUID)
func (r *EventLogRepository) GetByEventID(ctx context.Context, eventID string) (*database.EventLog, error) {
	event := new(database.EventLog)
	err := r.db.NewSelect().
		Model(event).
		Where("event_id = ?", eventID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// Update updates an existing event log
func (r *EventLogRepository) Update(ctx context.Context, event *database.EventLog) error {
	_, err := r.db.NewUpdate().
		Model(event).
		WherePK().
		Exec(ctx)
	return err
}

// Delete deletes an event log by ID
func (r *EventLogRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*database.EventLog)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// ListByTopic lists events by topic with pagination
func (r *EventLogRepository) ListByTopic(ctx context.Context, topic string, page, perPage int) ([]*database.EventLog, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	var events []*database.EventLog
	count, err := r.db.NewSelect().
		Model(&events).
		Where("topic = ?", topic).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		ScanAndCount(ctx)

	return events, count, err
}

// ListByScanID lists events for a specific scan (uses run_id column)
func (r *EventLogRepository) ListByScanID(ctx context.Context, scanID string) ([]*database.EventLog, error) {
	var events []*database.EventLog
	err := r.db.NewSelect().
		Model(&events).
		Where("run_id = ?", scanID).
		Order("created_at ASC").
		Scan(ctx)
	return events, err
}

// ListByWorkspace lists events for a workspace with pagination
func (r *EventLogRepository) ListByWorkspace(ctx context.Context, workspace string, page, perPage int) ([]*database.EventLog, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	var events []*database.EventLog
	count, err := r.db.NewSelect().
		Model(&events).
		Where("workspace = ?", workspace).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		ScanAndCount(ctx)

	return events, count, err
}

// ListUnprocessed lists unprocessed events
func (r *EventLogRepository) ListUnprocessed(ctx context.Context, limit int) ([]*database.EventLog, error) {
	if limit < 1 {
		limit = 100
	}

	var events []*database.EventLog
	err := r.db.NewSelect().
		Model(&events).
		Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Scan(ctx)
	return events, err
}

// MarkProcessed marks an event as processed
func (r *EventLogRepository) MarkProcessed(ctx context.Context, id int64, errorMsg string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*database.EventLog)(nil)).
		Set("processed = ?", true).
		Set("processed_at = ?", now).
		Set("error = ?", errorMsg).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// Search searches events with multiple criteria
func (r *EventLogRepository) Search(ctx context.Context, query EventLogQuery) ([]*database.EventLog, int, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 50
	}
	offset := (query.Page - 1) * query.PerPage

	var events []*database.EventLog
	q := r.db.NewSelect().Model(&events)

	if query.Topic != "" {
		q = q.Where("topic = ?", query.Topic)
	}
	if query.Name != "" {
		q = q.Where("name = ?", query.Name)
	}
	if query.Source != "" {
		q = q.Where("source = ?", query.Source)
	}
	if query.Workspace != "" {
		q = q.Where("workspace = ?", query.Workspace)
	}
	if query.ScanID != "" {
		q = q.Where("run_id = ?", query.ScanID)
	}
	if query.WorkflowName != "" {
		q = q.Where("workflow_name = ?", query.WorkflowName)
	}
	if query.Processed != nil {
		q = q.Where("processed = ?", *query.Processed)
	}
	if query.StartTime != nil {
		q = q.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		q = q.Where("created_at <= ?", *query.EndTime)
	}

	count, err := q.
		Order("created_at DESC").
		Limit(query.PerPage).
		Offset(offset).
		ScanAndCount(ctx)

	return events, count, err
}

// CountByTopic returns the count of events for a topic
func (r *EventLogRepository) CountByTopic(ctx context.Context, topic string) (int, error) {
	return r.db.NewSelect().
		Model((*database.EventLog)(nil)).
		Where("topic = ?", topic).
		Count(ctx)
}

// GetTopicSummary returns a summary of events by topic
func (r *EventLogRepository) GetTopicSummary(ctx context.Context) (map[string]int, error) {
	var results []struct {
		Topic string `bun:"topic"`
		Count int    `bun:"count"`
	}

	err := r.db.NewSelect().
		Model((*database.EventLog)(nil)).
		ColumnExpr("topic, COUNT(*) AS count").
		Group("topic").
		Order("count DESC").
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, r := range results {
		summary[r.Topic] = r.Count
	}

	return summary, nil
}

// DeleteOlderThan deletes events older than the specified time
func (r *EventLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*database.EventLog)(nil)).
		Where("created_at < ?", before).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteByWorkspace deletes all events for a workspace
func (r *EventLogRepository) DeleteByWorkspace(ctx context.Context, workspace string) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*database.EventLog)(nil)).
		Where("workspace = ?", workspace).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CreateBatch creates multiple event logs in a batch
func (r *EventLogRepository) CreateBatch(ctx context.Context, events []*database.EventLog) error {
	if len(events) == 0 {
		return nil
	}

	now := time.Now()
	for _, event := range events {
		if event.CreatedAt.IsZero() {
			event.CreatedAt = now
		}
	}

	_, err := r.db.NewInsert().Model(&events).Exec(ctx)
	return err
}

// GetRecentByWorkflow gets recent events for a workflow
func (r *EventLogRepository) GetRecentByWorkflow(ctx context.Context, workflowName string, limit int) ([]*database.EventLog, error) {
	if limit < 1 {
		limit = 10
	}

	var events []*database.EventLog
	err := r.db.NewSelect().
		Model(&events).
		Where("workflow_name = ?", workflowName).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	return events, err
}
