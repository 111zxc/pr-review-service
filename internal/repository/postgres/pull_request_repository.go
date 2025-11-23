package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/111zxc/pr-review-service/internal/domain"
)

type PullRequestRepository struct {
	pool *pgxpool.Pool
	tx   *TxManager
}

func NewPullRequestRepository(pool *pgxpool.Pool, tx *TxManager) *PullRequestRepository {
	return &PullRequestRepository{pool: pool, tx: tx}
}

func (r *PullRequestRepository) Create(pr *domain.PullRequest) error {
	ctx := context.Background()

	return r.tx.WithTx(ctx, func(tx pgx.Tx) error {
		statusQuery := `SELECT id FROM pr_statuses WHERE code = 'OPEN'`
		var statusID int
		if err := tx.QueryRow(ctx, statusQuery).Scan(&statusID); err != nil {
			return fmt.Errorf("failed to get OPEN status: %w", err)
		}

		prQuery := `
            INSERT INTO pull_requests (id, name, author_id, status_id, created_at)
            VALUES ($1, $2, $3, $4, $5)
        `
		if _, err := tx.Exec(ctx, prQuery, pr.ID, pr.Name, pr.AuthorID, statusID, time.Now()); err != nil {
			return fmt.Errorf("failed to create pull request: %w", err)
		}

		for _, reviewerID := range pr.AssignedReviewers {
			reviewerQuery := `INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)`
			if _, err := tx.Exec(ctx, reviewerQuery, pr.ID, reviewerID); err != nil {
				return fmt.Errorf("failed to add reviewer: %w", err)
			}
		}

		pr.Status = domain.PRStatusOpen
		now := time.Now()
		pr.CreatedAt = &now

		return nil
	})
}

func (r *PullRequestRepository) GetByID(id string) (*domain.PullRequest, error) {
	query := `
        SELECT 
            pr.id, pr.name, pr.author_id, 
            ps.code as status,
            pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pr_statuses ps ON pr.status_id = ps.id
        WHERE pr.id = $1
    `

	ctx := context.Background()
	var pr domain.PullRequest
	var statusCode string
	var mergedAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&statusCode,
		&pr.CreatedAt,
		&mergedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrPullRequestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	pr.Status = statusCode
	pr.MergedAt = mergedAt

	reviewers, err := r.getReviewers(ctx, id)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PullRequestRepository) Update(pr *domain.PullRequest) error {
	ctx := context.Background()

	return r.tx.WithTx(ctx, func(tx pgx.Tx) error {
		statusQuery := `SELECT id FROM pr_statuses WHERE code = $1`
		var statusID int
		if err := tx.QueryRow(ctx, statusQuery, pr.Status).Scan(&statusID); err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		query := `
            UPDATE pull_requests 
            SET name = $1, status_id = $2, updated_at = NOW(), merged_at = $3
            WHERE id = $4
        `
		result, err := tx.Exec(ctx, query, pr.Name, statusID, pr.MergedAt, pr.ID)
		if err != nil {
			return fmt.Errorf("failed to update pull request: %w", err)
		}

		if result.RowsAffected() == 0 {
			return domain.ErrPullRequestNotFound
		}

		if err := r.updateReviewers(ctx, tx, pr.ID, pr.AssignedReviewers); err != nil {
			return err
		}

		return nil
	})
}

func (r *PullRequestRepository) ListByReviewer(userID string) ([]*domain.PullRequest, error) {
	query := `
        SELECT 
            pr.id, pr.name, pr.author_id, 
            ps.code as status,
            pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pr_statuses ps ON pr.status_id = ps.id
        JOIN pr_reviewers prr ON pr.id = prr.pr_id
        WHERE prr.user_id = $1
        ORDER BY pr.created_at DESC
    `

	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var statusCode string
		var mergedAt *time.Time

		if err := rows.Scan(
			&pr.ID, &pr.Name, &pr.AuthorID, &statusCode, &pr.CreatedAt, &mergedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}

		pr.Status = statusCode
		pr.MergedAt = mergedAt

		reviewers, err := r.getReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers

		prs = append(prs, &pr)
	}

	return prs, nil
}

func (r *PullRequestRepository) Exists(id string) (bool, error) {
	query := `SELECT COUNT(*) FROM pull_requests WHERE id = $1`

	ctx := context.Background()
	var count int
	err := r.pool.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}

	return count > 0, nil
}

func (r *PullRequestRepository) getReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `SELECT user_id FROM pr_reviewers WHERE pr_id = $1`

	rows, err := r.pool.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, userID)
	}

	return reviewers, nil
}

func (r *PullRequestRepository) updateReviewers(ctx context.Context, tx pgx.Tx, prID string, reviewers []string) error {
	deleteQuery := `DELETE FROM pr_reviewers WHERE pr_id = $1`
	if _, err := tx.Exec(ctx, deleteQuery, prID); err != nil {
		return fmt.Errorf("failed to delete reviewers: %w", err)
	}

	for _, reviewerID := range reviewers {
		insertQuery := `INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)`
		if _, err := tx.Exec(ctx, insertQuery, prID, reviewerID); err != nil {
			return fmt.Errorf("failed to insert reviewer: %w", err)
		}
	}

	return nil
}
