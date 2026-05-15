package repository

import (
	"errors"

	"gorm.io/gorm"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
)

type MovieRepository struct {
	db *gorm.DB
}

func NewMovieRepository(db *gorm.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

func (r *MovieRepository) FindAll() ([]model.Movie, error) {
	var movies []model.Movie
	err := r.db.Find(&movies).Error
	return movies, err
}

func (r *MovieRepository) FindByID(id uint) (*model.Movie, error) {
	var movie model.Movie
	err := r.db.First(&movie, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &movie, err
}

func (r *MovieRepository) Create(movie *model.Movie) error {
	return r.db.Create(movie).Error
}

func (r *MovieRepository) Update(movie *model.Movie) error {
	return r.db.Save(movie).Error
}

func (r *MovieRepository) Delete(id uint) error {
	return r.db.Delete(&model.Movie{}, id).Error
}
