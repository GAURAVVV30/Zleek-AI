// Package auditlog writes and reads platform.audit_log entries (admin change
// audit and authentication events: who did what, when).
package auditlog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Record struct {
	ActorID          string
	Action           string
	TargetEntityType string
	TargetEntityID   string
	BeforeState      any
	AfterState       any
	CreatedAt        time.Time
}

type Writer interface {
	Write(ctx context.Context, rec Record) error
}

type Entry struct {
	ID               string          `json:"id"`
	ActorID          string          `json:"actor_id"`
	Action           string          `json:"action"`
	TargetEntityType string          `json:"target_entity_type"`
	TargetEntityID   string          `json:"target_entity_id"`
	BeforeState      json.RawMessage `json:"before_state,omitempty"`
	AfterState       json.RawMessage `json:"after_state,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

type PostgresAuditLog struct {
	db *pgxpool.Pool
}

func NewPostgresAuditLog(db *pgxpool.Pool) *PostgresAuditLog {
	return &PostgresAuditLog{db: db}
}

func (r *PostgresAuditLog) Write(ctx context.Context, rec Record) error {
	var before, after []byte
	var err error
	if rec.BeforeState != nil {
		if before, err = json.Marshal(rec.BeforeState); err != nil {
			return err
		}
	}
	if rec.AfterState != nil {
		if after, err = json.Marshal(rec.AfterState); err != nil {
			return err
		}
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	query := `
		INSERT INTO platform.audit_log
			(actor_id, action, target_entity_type, target_entity_id, before_state, after_state, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.Exec(ctx, query, rec.ActorID, rec.Action, rec.TargetEntityType, rec.TargetEntityID,
		nullableJSON(before), nullableJSON(after), rec.CreatedAt)
	return err
}

func (r *PostgresAuditLog) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT id, actor_id, action, target_entity_type, target_entity_id, before_state, after_state, created_at
		FROM platform.audit_log
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []Entry{}
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.TargetEntityType, &e.TargetEntityID,
			&e.BeforeState, &e.AfterState, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func nullableJSON(b []byte) any {
	if b == nil {
		return nil
	}
	return b
}
