package application

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hcl-backend/services/api-go/internal/identity/domain"
)

type AuthService struct {
	repo      UserRepository
	jwtSecret []byte
}

func NewAuthService(repo UserRepository, secret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	User         domain.User
}

func (s *AuthService) Signup(ctx context.Context, email, password string, role domain.UserRole) (*AuthResponse, error) {
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

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		Status:       domain.StatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

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

	return s.generateTokens(user)
}

func (s *AuthService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AuthService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

func (s *AuthService) UpdateUser(ctx context.Context, id string, role *string, status *string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != nil {
		user.Role = domain.UserRole(*role)
	}
	if status != nil {
		user.Status = domain.UserStatus(*status)
	}
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) generateTokens(user *domain.User) (*AuthResponse, error) {
	// Access Token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"type": "access",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh Token
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
