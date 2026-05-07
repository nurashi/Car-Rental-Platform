package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nurashi/car-rental-identity/internal/domain"
	"github.com/nurashi/car-rental-identity/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUserBlocked        = errors.New("user account is blocked")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type Config struct {
	JWTSecret      string
	JWTExpiryHours int
}

type AuthService struct {
	userRepo     repository.UserRepository
	emailVerRepo repository.EmailVerificationRepository
	config       Config
}

func NewAuthService(userRepo repository.UserRepository, emailVerRepo repository.EmailVerificationRepository, config Config) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailVerRepo: emailVerRepo,
		config:       config,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName, phone string) (string, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return "", ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
		Status:       domain.UserStatusPending,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	token, err := generateVerificationToken()
	if err != nil {
		return "", fmt.Errorf("generate verification token: %w", err)
	}

	ev := &domain.EmailVerification{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.emailVerRepo.Create(ctx, ev); err != nil {
		return "", fmt.Errorf("create email verification: %w", err)
	}

	return user.ID, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if user.Status == domain.UserStatusBlocked {
		return nil, "", ErrUserBlocked
	}

	token, err := s.generateJWT(user.ID, user.Email)
	if err != nil {
		return nil, "", fmt.Errorf("generate jwt: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	ev, err := s.emailVerRepo.GetByToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	if ev.Used {
		return errors.New("verification token already used")
	}

	if time.Now().After(ev.ExpiresAt) {
		return errors.New("verification token expired")
	}

	if err := s.userRepo.UpdateEmailVerified(ctx, ev.UserID, true); err != nil {
		return fmt.Errorf("update email verified: %w", err)
	}

	if err := s.emailVerRepo.MarkUsed(ctx, ev.ID); err != nil {
		return fmt.Errorf("mark verification used: %w", err)
	}

	if err := s.userRepo.Update(ctx, &domain.User{ID: ev.UserID, Status: domain.UserStatusActive}); err != nil {
		return fmt.Errorf("update user status: %w", err)
	}

	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return ErrInvalidCredentials
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	_ = s.emailVerRepo.DeleteByUserID(ctx, user.ID)

	token, err := generateVerificationToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	ev := &domain.EmailVerification{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.emailVerRepo.Create(ctx, ev); err != nil {
		return fmt.Errorf("create email verification: %w", err)
	}

	return nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return nil
}

func (s *AuthService) generateJWT(userID, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     now.Unix(),
		"exp":     now.Add(time.Duration(s.config.JWTExpiryHours) * time.Hour).Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.config.JWTSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

func generateVerificationToken() (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type": "verification",
		"iat":  time.Now().Unix(),
	}).SignedString([]byte("verification-secret"))
	if err != nil {
		return "", fmt.Errorf("sign verification token: %w", err)
	}
	return token, nil
}

func HashPasswordForTest(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthService) GenerateJWTForTest(userID, email string) (string, error) {
	return s.generateJWT(userID, email)
}
