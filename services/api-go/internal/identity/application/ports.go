package application

import (
	"context"

	"github.com/hcl-backend/services/api-go/internal/identity/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	List(ctx context.Context) ([]domain.User, error)
}
