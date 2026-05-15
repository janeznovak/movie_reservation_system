package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
)

type ShowtimeRepository struct {
	db *gorm.DB
}

func NewShowtimeRepository(db *gorm.DB) *ShowtimeRepository {
	return &ShowtimeRepository{db: db}
}

func (r *ShowtimeRepository) FindByID(id uint) (*model.Showtime, error) {
	var showtime model.Showtime
	err := r.db.Preload("Movie").Preload("Venue").First(&showtime, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &showtime, err
}

func (r *ShowtimeRepository) FindByDate(date time.Time) ([]model.Showtime, error) {
	var showtimes []model.Showtime
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	err := r.db.Preload("Movie").Preload("Venue").
		Where("start_time >= ? AND start_time < ?", start, end).
		Order("start_time").
		Find(&showtimes).Error

	return showtimes, err
}

func (r *ShowtimeRepository) FindByMovieID(movieID uint) ([]model.Showtime, error) {
	var showtimes []model.Showtime
	err := r.db.Preload("Movie").Preload("Venue").
		Where("movie_id = ?", movieID).
		Order("start_time").
		Find(&showtimes).Error
	return showtimes, err
}

func (r *ShowtimeRepository) Create(showtime *model.Showtime) error {
	return r.db.Create(showtime).Error
}

func (r *ShowtimeRepository) Update(showtime *model.Showtime) error {
	return r.db.Save(showtime).Error
}

func (r *ShowtimeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Showtime{}, id).Error
}
