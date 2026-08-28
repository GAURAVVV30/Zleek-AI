package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/admin/domain"
)

type IdentityService interface {
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, id string, role, status *string) (*domain.User, error)
}

type AuditRepository interface {
	Create(ctx context.Context, record *domain.AuditRecord) error
	List(ctx context.Context) ([]domain.AuditRecord, error)
}
