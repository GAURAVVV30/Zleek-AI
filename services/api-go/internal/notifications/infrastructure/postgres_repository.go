package infrastructure

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/hcl-backend/services/api-go/internal/notifications/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO platform.notifications (learner_id, event_type, payload)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, n.LearnerID, n.EventType, n.Payload).Scan(&n.ID, &n.CreatedAt)
}

func (r *PostgresRepository) ListForUser(ctx context.Context, learnerID string) ([]domain.Notification, error) {
	query := `
		SELECT id, learner_id, event_type, payload, read_at, created_at
		FROM platform.notifications
		WHERE learner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, learnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.LearnerID, &n.EventType, &n.Payload, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	query := `
		SELECT id, learner_id, event_type, payload, read_at, created_at
		FROM platform.notifications
		WHERE id = $1
	`
	var n domain.Notification
	err := r.db.QueryRow(ctx, query, id).Scan(&n.ID, &n.LearnerID, &n.EventType, &n.Payload, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, domain.ErrNotificationNotFound
	}
	return &n, nil
}

func (r *PostgresRepository) MarkRead(ctx context.Context, id string) error {
	query := `
		UPDATE platform.notifications
		SET read_at = $1
		WHERE id = $2 AND read_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}
