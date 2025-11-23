package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/111zxc/pr-review-service/internal/domain"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) GetEventStats() (*domain.StatsResponse, error) {
	query := `
        SELECT 
            event_type,
            COUNT(*) as count
        FROM events
        GROUP BY event_type
        ORDER BY count DESC
    `

	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := &domain.StatsResponse{
		EventCounts: make(map[domain.EventType]int),
		TotalEvents: 0,
	}

	for rows.Next() {
		var eventType string
		var count int

		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, err
		}

		stats.EventCounts[domain.EventType(eventType)] = count
		stats.TotalEvents += count
	}

	allEventTypes := []domain.EventType{
		domain.EventTypePRCreated,
		domain.EventTypePRMerged,
		domain.EventTypeReviewerAssigned,
		domain.EventTypeReviewerReassigned,
		domain.EventTypeReviewerUnassigned,
	}

	for _, eventType := range allEventTypes {
		if _, exists := stats.EventCounts[eventType]; !exists {
			stats.EventCounts[eventType] = 0
		}
	}

	return stats, nil
}
