package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/identity/domain"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

const userColumns = `id, email, password_hash, role, status, full_name, timezone, theme, created_at, updated_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.FullName, &user.Timezone, &user.Theme, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO platform.users (id, email, password_hash, role, status, full_name, timezone, theme, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.Status,
		user.FullName, user.Timezone, user.Theme, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM platform.users WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM platform.users WHERE email = $1`, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE platform.users
		SET email = $1, password_hash = $2, role = $3, status = $4, full_name = $5,
			timezone = $6, theme = $7, updated_at = $8
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query, user.Email, user.PasswordHash, user.Role, user.Status,
		user.FullName, user.Timezone, user.Theme, user.UpdatedAt, user.ID)
	return err
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM platform.users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, rows.Err()
}
