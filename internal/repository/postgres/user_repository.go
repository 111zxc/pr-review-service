package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/111zxc/pr-review-service/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
        INSERT INTO users (id, username, is_active) 
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            is_active = EXCLUDED.is_active,
            updated_at = NOW()
    `

	ctx := context.Background()
	_, err := r.pool.Exec(ctx, query, user.ID, user.Username, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	query := `
        SELECT u.id, u.username, u.is_active, t.name as team_name
        FROM users u
        LEFT JOIN team_members tm ON u.id = tm.user_id
        LEFT JOIN teams t ON tm.team_id = t.id
        WHERE u.id = $1 AND u.deleted_at IS NULL
    `

	ctx := context.Background()
	var user domain.User
	var teamName *string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.IsActive, &teamName,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if teamName != nil {
		user.TeamName = *teamName
	}

	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	query := `
        UPDATE users 
        SET username = $1, is_active = $2, updated_at = NOW()
        WHERE id = $3 AND deleted_at IS NULL
    `

	ctx := context.Background()
	result, err := r.pool.Exec(ctx, query, user.Username, user.IsActive, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) GetByTeam(teamName string) ([]*domain.User, error) {
	query := `
        SELECT u.id, u.username, u.is_active
        FROM users u
        JOIN team_members tm ON u.id = tm.user_id
        JOIN teams t ON tm.team_id = t.id
        WHERE t.name = $1 AND u.deleted_at IS NULL AND u.is_active = true
    `

	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		user.TeamName = teamName
		users = append(users, &user)
	}

	return users, nil
}
