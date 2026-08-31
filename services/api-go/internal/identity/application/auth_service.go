package application

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hcl-backend/services/api-go/internal/identity/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/auditlog"
)

type AuthService struct {
	repo      UserRepository
	audit     auditlog.Writer
	jwtSecret []byte
}

func NewAuthService(repo UserRepository, secret string, audit auditlog.Writer) *AuthService {
	return &AuthService{
		repo:      repo,
		audit:     audit,
		jwtSecret: []byte(secret),
	}
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	User         domain.User
}

func (s *AuthService) Signup(ctx context.Context, email, password, fullName string) (*AuthResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !domain.ValidateGmail(email) {
		return nil, domain.ErrInvalidEmail
	}
	if !domain.ValidatePassword(password) {
		return nil, domain.ErrWeakPassword
	}

	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if err != domain.ErrUserNotFound {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleLearner,
		Status:       domain.StatusActive,
		FullName:     strings.TrimSpace(fullName),
		Timezone:     "UTC",
		Theme:        "default",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.auditLoginEvent(ctx, user, "auth.signup", user)
	return s.generateTokens(user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status == domain.StatusSuspended {
		return nil, domain.ErrInvalidCredentials
	}

	s.auditLoginEvent(ctx, user, "auth.login", user)
	return s.generateTokens(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		return nil, domain.ErrInvalidCredentials
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	s.auditLoginEvent(ctx, user, "auth.refresh", user)
	return s.generateTokens(user)
}

func (s *AuthService) Logout(ctx context.Context, actorID string) error {
	if s.audit == nil {
		return nil
	}
	return s.audit.Write(ctx, auditlog.Record{
		ActorID:          actorID,
		Action:           "auth.logout",
		TargetEntityType: "user",
		TargetEntityID:   actorID,
	})
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, current, newPassword string) error {
	if !domain.ValidatePassword(newPassword) {
		return domain.ErrWeakPassword
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(current)) != nil {
		return domain.ErrInvalidCredentials
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashed)
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return err
	}
	s.auditLoginEvent(ctx, user, "auth.change_password", user)
	return nil
}

func (s *AuthService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AuthService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *AuthService) UpdateUser(ctx context.Context, id string, role *string, status *string, fullName, timezone, theme *string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	before := map[string]any{"role": user.Role, "status": user.Status}
	if role != nil {
		user.Role = domain.UserRole(*role)
	}
	if status != nil {
		user.Status = domain.UserStatus(*status)
	}
	if fullName != nil {
		user.FullName = *fullName
	}
	if timezone != nil {
		user.Timezone = *timezone
	}
	if theme != nil {
		user.Theme = *theme
	}
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	after := map[string]any{"role": user.Role, "status": user.Status}
	s.auditLoginEvent(ctx, user, "admin.update_user", map[string]any{"before": before, "after": after})
	return user, nil
}

func (s *AuthService) Initialize(ctx context.Context, user *domain.User) error {
	return s.repo.Create(ctx, user)
}

func (s *AuthService) auditLoginEvent(ctx context.Context, user *domain.User, action string, after any) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Write(ctx, auditlog.Record{
		ActorID:          user.ID,
		Action:           action,
		TargetEntityType: "user",
		TargetEntityID:   user.ID,
		AfterState:       after,
	})
}

func (s *AuthService) generateTokens(user *domain.User) (*AuthResponse, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"type": "access",
		"exp":  time.Now().Add(7 * time.Hour).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"type": "refresh",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		User:         *user,
	}, nil
}
