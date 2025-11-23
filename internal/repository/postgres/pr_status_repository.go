package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/111zxc/pr-review-service/internal/domain"
)

type PRStatusRepository struct {
	pool *pgxpool.Pool
}

func NewPRStatusRepository(pool *pgxpool.Pool) *PRStatusRepository {
	return &PRStatusRepository{pool: pool}
}

func (r *PRStatusRepository) GetByCode(code string) (*domain.PRStatus, error) {
	query := `
        SELECT id, code, name, description, created_at 
        FROM pr_statuses 
        WHERE code = $1
    `

	ctx := context.Background()
	var status domain.PRStatus
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&status.ID,
		&status.Code,
		&status.Name,
		&status.Description,
		&status.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("PR status not found: %s", code)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PR status: %w", err)
	}

	return &status, nil
}

func (r *PRStatusRepository) GetByID(id int) (*domain.PRStatus, error) {
	query := `
        SELECT id, code, name, description, created_at 
        FROM pr_statuses 
        WHERE id = $1
    `

	ctx := context.Background()
	var status domain.PRStatus
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&status.ID,
		&status.Code,
		&status.Name,
		&status.Description,
		&status.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("PR status not found with ID: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PR status: %w", err)
	}

	return &status, nil
}

func (r *PRStatusRepository) ListAll() ([]*domain.PRStatus, error) {
	query := `SELECT id, code, name, description, created_at FROM pr_statuses ORDER BY id`

	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query PR statuses: %w", err)
	}
	defer rows.Close()

	var statuses []*domain.PRStatus
	for rows.Next() {
		var status domain.PRStatus
		if err := rows.Scan(&status.ID, &status.Code, &status.Name, &status.Description, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan PR status: %w", err)
		}
		statuses = append(statuses, &status)
	}

	return statuses, nil
}
