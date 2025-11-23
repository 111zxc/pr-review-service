package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/111zxc/pr-review-service/internal/domain"
)

type EventsRepository struct {
	pool *pgxpool.Pool
}

func NewEventsRepository(pool *pgxpool.Pool) *EventsRepository {
	return &EventsRepository{pool: pool}
}

func (r *EventsRepository) CreateEvent(event *domain.Event) error {
	query := `
        INSERT INTO events (event_type, pr_id, user_id, additional_data)
        VALUES ($1, $2, $3, $4)
    `

	ctx := context.Background()
	_, err := r.pool.Exec(ctx, query, event.EventType, event.PRID, event.UserID, event.AdditionalData)
	if err != nil {
		return err
	}

	return nil
}

func (r *EventsRepository) GetEventsByType(eventType domain.EventType, limit int) ([]domain.Event, error) {
	query := `
        SELECT id, event_type, pr_id, user_id, additional_data, created_at
        FROM events
        WHERE event_type = $1
        ORDER BY created_at DESC
        LIMIT $2
    `

	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		err := rows.Scan(&event.ID, &event.EventType, &event.PRID, &event.UserID, &event.AdditionalData, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventsRepository) GetEventCountsByType() (map[domain.EventType]int, error) {
	query := `
        SELECT event_type, COUNT(*) 
        FROM events 
        GROUP BY event_type
    `

	ctx := context.Background()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[domain.EventType]int)
	for rows.Next() {
		var eventType string
		var count int
		err := rows.Scan(&eventType, &count)
		if err != nil {
			return nil, err
		}
		counts[domain.EventType(eventType)] = count
	}

	return counts, nil
}
