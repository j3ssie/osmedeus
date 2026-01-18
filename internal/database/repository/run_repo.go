package repository

import (
	"context"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/uptrace/bun"
)

// RunRepository handles run database operations
type RunRepository struct {
	db *bun.DB
}

// NewRunRepository creates a new run repository
func NewRunRepository(db *bun.DB) *RunRepository {
	return &RunRepository{db: db}
}

// Create creates a new run
func (r *RunRepository) Create(ctx context.Context, scan *database.Run) error {
	_, err := r.db.NewInsert().Model(scan).Exec(ctx)
	return err
}

// GetByID gets a run by ID
func (r *RunRepository) GetByID(ctx context.Context, id string) (*database.Run, error) {
	scan := new(database.Run)
	err := r.db.NewSelect().
		Model(scan).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scan, nil
}

// GetByRunID gets a run by run ID
func (r *RunRepository) GetByRunID(ctx context.Context, runID string) (*database.Run, error) {
	scan := new(database.Run)
	err := r.db.NewSelect().
		Model(scan).
		Where("run_id = ?", runID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scan, nil
}

// Update updates a run
func (r *RunRepository) Update(ctx context.Context, scan *database.Run) error {
	_, err := r.db.NewUpdate().
		Model(scan).
		WherePK().
		Exec(ctx)
	return err
}

// Delete deletes a run by ID
func (r *RunRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*database.Run)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// List lists runs with pagination
func (r *RunRepository) List(ctx context.Context, page, perPage int) ([]*database.Run, int, error) {
	var scans []*database.Run

	count, err := r.db.NewSelect().
		Model(&scans).
		Order("created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	return scans, count, nil
}

// ListByStatus lists runs by status
func (r *RunRepository) ListByStatus(ctx context.Context, status string) ([]*database.Run, error) {
	var scans []*database.Run
	err := r.db.NewSelect().
		Model(&scans).
		Where("status = ?", status).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scans, nil
}

// ListByWorkflow lists runs by workflow name
func (r *RunRepository) ListByWorkflow(ctx context.Context, workflowName string) ([]*database.Run, error) {
	var scans []*database.Run
	err := r.db.NewSelect().
		Model(&scans).
		Where("workflow_name = ?", workflowName).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scans, nil
}

// ListByTarget lists runs by target
func (r *RunRepository) ListByTarget(ctx context.Context, target string) ([]*database.Run, error) {
	var scans []*database.Run
	err := r.db.NewSelect().
		Model(&scans).
		Where("target = ?", target).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scans, nil
}

// GetWithSteps gets a run with its step results
func (r *RunRepository) GetWithSteps(ctx context.Context, id string) (*database.Run, error) {
	scan := new(database.Run)
	err := r.db.NewSelect().
		Model(scan).
		Relation("Steps").
		Where("r.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scan, nil
}

// GetWithArtifacts gets a run with its artifacts
func (r *RunRepository) GetWithArtifacts(ctx context.Context, id string) (*database.Run, error) {
	scan := new(database.Run)
	err := r.db.NewSelect().
		Model(scan).
		Relation("Artifacts").
		Where("r.id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return scan, nil
}
