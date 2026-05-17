package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
)

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Create converts a set of seat holds into a confirmed reservation inside a transaction.
// Holds are locked with SELECT FOR UPDATE, validated, then deleted once the reservation is inserted.
func (r *ReservationRepository) Create(userID uint, holdIDs []uint) (*model.Reservation, error) {
	var reservation model.Reservation

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var holds []model.SeatHold
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", holdIDs).
			Find(&holds).Error; err != nil {
			return err
		}

		if len(holds) != len(holdIDs) {
			return apperr.ErrNotFound
		}

		showtimeID := holds[0].ShowtimeID
		now := time.Now()
		for _, h := range holds {
			if h.UserID != userID {
				return apperr.ErrNotFound
			}
			if h.ShowtimeID != showtimeID {
				return errors.New("all holds must be for the same showtime")
			}
			if !h.ExpiresAt.After(now) {
				return apperr.ErrExpired
			}
		}

		var showtime model.Showtime
		if err := tx.First(&showtime, showtimeID).Error; err != nil {
			return err
		}

		reservation = model.Reservation{
			UserID:     userID,
			ShowtimeID: showtimeID,
			Status:     model.StatusConfirmed,
			TotalPrice: showtime.Price * float64(len(holds)),
		}
		if err := tx.Create(&reservation).Error; err != nil {
			return err
		}

		var seatIDs []uint
		for _, h := range holds {
			seatIDs = append(seatIDs, h.SeatID)
		}
		var seats []model.Seat
		if err := tx.Where("id IN ?", seatIDs).Find(&seats).Error; err != nil {
			return err
		}
		if err := tx.Model(&reservation).Association("Seats").Append(seats); err != nil {
			return err
		}

		return tx.Delete(&model.SeatHold{}, holdIDs).Error
	})

	if err != nil {
		return nil, err
	}
	return r.FindByID(reservation.ID)
}

func (r *ReservationRepository) FindByID(id uint) (*model.Reservation, error) {
	var reservation model.Reservation
	err := r.db.
		Preload("Showtime", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Movie").Preload("Venue")
		}).
		Preload("Seats").
		First(&reservation, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	return &reservation, err
}

func (r *ReservationRepository) FindByUserID(userID uint) ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := r.db.
		Preload("Showtime", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Movie").Preload("Venue")
		}).
		Preload("Seats").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reservations).Error
	return reservations, err
}

func (r *ReservationRepository) Cancel(id uint) error {
	return r.db.Model(&model.Reservation{}).
		Where("id = ?", id).
		Update("status", model.StatusCancelled).Error
}

func (r *ReservationRepository) FindAll() ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := r.db.
		Preload("User").
		Preload("Showtime", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Movie").Preload("Venue")
		}).
		Preload("Seats").
		Order("created_at DESC").
		Find(&reservations).Error
	return reservations, err
}

func (r *ReservationRepository) GetReport() ([]model.ShowtimeReport, error) {
	var reports []model.ShowtimeReport
	err := r.db.Raw(`
		SELECT
			st.id                                                                    AS showtime_id,
			m.title                                                                  AS movie,
			v.name                                                                   AS venue,
			st.start_time,
			(SELECT COUNT(*) FROM seats
			 WHERE venue_id = st.venue_id AND deleted_at IS NULL)                   AS total_seats,
			(SELECT COUNT(DISTINCT rs.seat_id)
			 FROM reservation_seats rs
			 JOIN reservations r ON r.id = rs.reservation_id
			 WHERE r.showtime_id = st.id AND r.status = 'confirmed'
			   AND r.deleted_at IS NULL)                                             AS seats_sold,
			(SELECT COUNT(*) FROM seats
			 WHERE venue_id = st.venue_id AND deleted_at IS NULL) -
			(SELECT COUNT(DISTINCT rs.seat_id)
			 FROM reservation_seats rs
			 JOIN reservations r ON r.id = rs.reservation_id
			 WHERE r.showtime_id = st.id AND r.status = 'confirmed'
			   AND r.deleted_at IS NULL)                                             AS seats_available,
			(SELECT COALESCE(SUM(total_price), 0)
			 FROM reservations r
			 WHERE r.showtime_id = st.id AND r.status = 'confirmed'
			   AND r.deleted_at IS NULL)                                             AS revenue
		FROM showtimes st
		JOIN movies m ON m.id = st.movie_id AND m.deleted_at IS NULL
		JOIN venues v ON v.id = st.venue_id AND v.deleted_at IS NULL
		WHERE st.deleted_at IS NULL
		ORDER BY st.start_time
	`).Scan(&reports).Error
	return reports, err
}
