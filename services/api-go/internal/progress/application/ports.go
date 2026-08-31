package application

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/hcl-backend/services/api-go/internal/progress/domain"
)

// TxManager wraps a database transaction so use cases stay infrastructure-free.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

type ProgressRepository interface {
	RecordEvidence(ctx context.Context, tx pgx.Tx, evidence *domain.Evidence) error
	RecordEngagement(ctx context.Context, event *domain.EngagementEvent) error
	SyncPathItemState(ctx context.Context, tx pgx.Tx, learnerID, conceptNodeID, state string) error
	Summary(ctx context.Context, learnerID, structureID string) (*SummaryPayload, error)
	CompetentConceptNames(ctx context.Context, learnerID, structureID string) ([]string, error)
}

// SummaryPayload is the raw aggregate the repository computes.
type SummaryPayload struct {
	TotalConcepts int
	Competent     int
	InProgress    int
	Weak          int
	Remediations  int
	Breakdown     []domain.SummaryRow
}

// CompetencyService is the application-level port to update competency state
// deterministically within the caller transaction.
type CompetencyService interface {
	UpdateWithEvidence(ctx context.Context, tx pgx.Tx, learnerID, conceptID, result, evidenceID string) error
}

// GoalService resolves the learner's active goal structure for scoping the
// progress dashboard.
type GoalService interface {
	ActiveStructureMeta(ctx context.Context, learnerID string) (goalID string, goalTitle string, structureID string, err error)
}
