package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/111zxc/pr-review-service/internal/domain"
)

type TeamRepository struct {
	pool *pgxpool.Pool
	tx   *TxManager
}

func NewTeamRepository(pool *pgxpool.Pool, tx *TxManager) *TeamRepository {
	return &TeamRepository{pool: pool, tx: tx}
}

func (r *TeamRepository) Create(team *domain.Team) error {
	ctx := context.Background()

	return r.tx.WithTx(ctx, func(tx pgx.Tx) error {
		teamQuery := `INSERT INTO teams (id, name) VALUES ($1, $2)`
		if _, err := tx.Exec(ctx, teamQuery, team.Name, team.Name); err != nil {
			return fmt.Errorf("failed to create team: %w", err)
		}

		for _, member := range team.Members {
			memberQuery := `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`
			if _, err := tx.Exec(ctx, memberQuery, team.Name, member.UserID); err != nil {
				return fmt.Errorf("failed to add team member: %w", err)
			}
		}

		return nil
	})
}

func (r *TeamRepository) GetByName(name string) (*domain.Team, error) {
	ctx := context.Background()

	teamQuery := `SELECT name FROM teams WHERE name = $1 AND deleted_at IS NULL`

	var team domain.Team
	err := r.pool.QueryRow(ctx, teamQuery, name).Scan(&team.Name)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	membersQuery := `
        SELECT u.id, u.username, u.is_active
        FROM users u
        JOIN team_members tm ON u.id = tm.user_id
        WHERE tm.team_id = $1 AND u.deleted_at IS NULL
    `

	rows, err := r.pool.Query(ctx, membersQuery, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		team.Members = append(team.Members, member)
	}

	return &team, nil
}

func (r *TeamRepository) Exists(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM teams WHERE name = $1 AND deleted_at IS NULL`

	ctx := context.Background()
	var count int
	err := r.pool.QueryRow(ctx, query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return count > 0, nil
}
