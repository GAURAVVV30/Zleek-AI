package infrastructure

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/admin/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuditRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAuditRepository(db *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{db: db}
}

func (r *PostgresAuditRepository) Create(ctx context.Context, record *domain.AuditRecord) error {
	query := `
		INSERT INTO platform.audit_log (actor_id, action, target_entity_type, target_entity_id, before_state, after_state)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		record.ActorID,
		record.Action,
		record.TargetEntityType,
		record.TargetEntityID,
		record.BeforeState,
		record.AfterState,
	).Scan(&record.ID, &record.CreatedAt)
}

func (r *PostgresAuditRepository) List(ctx context.Context) ([]domain.AuditRecord, error) {
	query := `
		SELECT id, actor_id, action, target_entity_type, target_entity_id, before_state, after_state, created_at
		FROM platform.audit_log
		ORDER BY created_at DESC
		LIMIT 100
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.AuditRecord
	for rows.Next() {
		var rec domain.AuditRecord
		if err := rows.Scan(
			&rec.ID,
			&rec.ActorID,
			&rec.Action,
			&rec.TargetEntityType,
			&rec.TargetEntityID,
			&rec.BeforeState,
			&rec.AfterState,
			&rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
