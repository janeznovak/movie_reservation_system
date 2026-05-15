package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
)

type VenueRepository struct {
	db *gorm.DB
}

func NewVenueRepository(db *gorm.DB) *VenueRepository {
	return &VenueRepository{db: db}
}

func (r *VenueRepository) FindAll() ([]model.Venue, error) {
	var venues []model.Venue
	err := r.db.Find(&venues).Error
	return venues, err
}

func (r *VenueRepository) FindByID(id uint) (*model.Venue, error) {
	var venue model.Venue
	err := r.db.First(&venue, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &venue, err
}

func (r *VenueRepository) Create(venue *model.Venue) error {
	return r.db.Create(venue).Error
}

func (r *VenueRepository) FindSeatsByVenueID(venueID uint) ([]model.Seat, error) {
	var seats []model.Seat
	err := r.db.Where("venue_id = ?", venueID).Find(&seats).Error
	return seats, err
}

func (r *VenueRepository) FindSeatByID(id uint) (*model.Seat, error) {
	var seat model.Seat
	err := r.db.First(&seat, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &seat, err
}

func (r *VenueRepository) CreateSeat(seat *model.Seat) error {
	err := r.db.Create(seat).Error
	if err != nil {
		// Detect PostgreSQL unique constraint violation (e.g. duplicate row+number in same venue)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperr.ErrConflict
		}
	}
	return err
}

func (r *VenueRepository) DeleteSeat(id uint) error {
	return r.db.Delete(&model.Seat{}, id).Error
}
