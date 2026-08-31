package application

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

type ProgressRepository interface {
	RecordEvidence(ctx context.Context, tx pgx.Tx, evidence *domain.Evidence) error
	RecordEngagement(ctx context.Context, event *domain.EngagementEvent) error
	GetCompetentCount(ctx context.Context, learnerID string) (int, error)
	GetInProgressCount(ctx context.Context, learnerID string) (int, error)
}

// CompetencyService is the application-level port to update competency state deterministically.
type CompetencyService interface {
	UpdateWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, conceptID, result, evidenceID string) error
}
