package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type Claims struct {
	UserID uint       `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(name, email, password string) (string, error) {
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return "", apperr.ErrConflict
	}
	if !errors.Is(err, apperr.ErrNotFound) {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", err
	}

	return s.generateToken(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Return a generic message so we don't reveal whether the email exists
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.generateToken(user)
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
