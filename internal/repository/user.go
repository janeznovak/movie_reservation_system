package repository

import (
	"errors"

	"gorm.io/gorm"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &user, err
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}
