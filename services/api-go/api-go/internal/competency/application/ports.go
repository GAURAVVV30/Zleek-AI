package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/competency/domain"
	"github.com/jackc/pgx/v5"
)

type CompetencyRepository interface {
	GetByLearnerAndConcept(ctx context.Context, learnerID, conceptID string) (*domain.CompetencyRecord, error)
	GetHistoryByLearnerAndConcept(ctx context.Context, learnerID, conceptID string) ([]domain.CompetencyHistory, error)
	UpsertCompetency(ctx context.Context, tx pgx.Tx, record *domain.CompetencyRecord) error
	AppendHistory(ctx context.Context, tx pgx.Tx, history *domain.CompetencyHistory) error
}

type ConceptService interface {
	ValidateConcept(ctx context.Context, conceptID string) error
}
