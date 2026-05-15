package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
)

type SeatHoldRepository struct {
	db *gorm.DB
}

func NewSeatHoldRepository(db *gorm.DB) *SeatHoldRepository {
	return &SeatHoldRepository{db: db}
}

func (r *SeatHoldRepository) FindAvailableSeats(venueID, showtimeID uint) ([]model.Seat, error) {
	var seats []model.Seat
	err := r.db.Where(`
		venue_id = ?
		AND id NOT IN (
			SELECT rs.seat_id FROM reservation_seats rs
			JOIN reservations ON reservations.id = rs.reservation_id
			WHERE reservations.showtime_id = ?
			  AND reservations.status = 'confirmed'
			  AND reservations.deleted_at IS NULL
		)
		AND id NOT IN (
			SELECT sh.seat_id FROM seat_holds sh
			WHERE sh.showtime_id = ?
			  AND sh.expires_at > ?
			  AND sh.deleted_at IS NULL
		)
	`, venueID, showtimeID, showtimeID, time.Now()).
		Order("row_label, seat_number").
		Find(&seats).Error
	return seats, err
}

// CreateHolds opens a transaction, locks the requested seat rows with SELECT FOR UPDATE,
// checks no seat is already held or reserved, then inserts a hold per seat.
func (r *SeatHoldRepository) CreateHolds(showtimeID, userID uint, seatIDs []uint) ([]model.SeatHold, error) {
	var holds []model.SeatHold

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Locking the seat rows stops a concurrent request from passing the same availability
		// check before either transaction has inserted its holds.
		var seats []model.Seat
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", seatIDs).
			Find(&seats).Error; err != nil {
			return err
		}

		var holdCount int64
		if err := tx.Model(&model.SeatHold{}).
			Where("seat_id IN ? AND showtime_id = ? AND expires_at > ? AND deleted_at IS NULL",
				seatIDs, showtimeID, time.Now()).
			Count(&holdCount).Error; err != nil {
			return err
		}
		if holdCount > 0 {
			return apperr.ErrConflict
		}

		var reservedCount int64
		if err := tx.Table("reservation_seats").
			Joins("JOIN reservations ON reservations.id = reservation_seats.reservation_id").
			Where("reservation_seats.seat_id IN ? AND reservations.showtime_id = ? AND reservations.deleted_at IS NULL AND reservations.status = 'confirmed'",
				seatIDs, showtimeID).
			Count(&reservedCount).Error; err != nil {
			return err
		}
		if reservedCount > 0 {
			return apperr.ErrConflict
		}

		expiresAt := time.Now().Add(10 * time.Minute)
		for _, seatID := range seatIDs {
			hold := model.SeatHold{
				SeatID:     seatID,
				ShowtimeID: showtimeID,
				UserID:     userID,
				ExpiresAt:  expiresAt,
			}
			if err := tx.Create(&hold).Error; err != nil {
				return err
			}
			holds = append(holds, hold)
		}

		var holdIDs []uint
		for _, h := range holds {
			holdIDs = append(holdIDs, h.ID)
		}
		var full []model.SeatHold
		if err := tx.Preload("Seat").Where("id IN ?", holdIDs).Find(&full).Error; err != nil {
			return err
		}
		holds = full
		return nil
	})

	return holds, err
}

func (r *SeatHoldRepository) FindByID(id uint) (*model.SeatHold, error) {
	var hold model.SeatHold
	err := r.db.Preload("Seat").First(&hold, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &hold, err
}

func (r *SeatHoldRepository) Delete(id uint) error {
	return r.db.Delete(&model.SeatHold{}, id).Error
}
