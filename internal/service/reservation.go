package service

import (
	"errors"
	"time"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type ReservationService struct {
	reservationRepo *repository.ReservationRepository
}

func NewReservationService(reservationRepo *repository.ReservationRepository) *ReservationService {
	return &ReservationService{reservationRepo: reservationRepo}
}

func (s *ReservationService) CreateReservation(userID uint, holdIDs []uint) (*model.Reservation, error) {
	if len(holdIDs) == 0 {
		return nil, errors.New("at least one hold is required")
	}

	seen := make(map[uint]bool)
	unique := make([]uint, 0, len(holdIDs))
	for _, id := range holdIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return s.reservationRepo.Create(userID, unique)
}

func (s *ReservationService) ListReservations(userID uint) ([]model.Reservation, error) {
	return s.reservationRepo.FindByUserID(userID)
}

func (s *ReservationService) CancelReservation(reservationID, userID uint) error {
	reservation, err := s.reservationRepo.FindByID(reservationID)
	if err != nil {
		return err
	}

	if reservation.UserID != userID {
		return apperr.ErrNotFound
	}

	if reservation.Status == model.StatusCancelled {
		return errors.New("reservation is already cancelled")
	}

	if !reservation.Showtime.StartTime.After(time.Now()) {
		return errors.New("cannot cancel a reservation after the showtime has started")
	}

	return s.reservationRepo.Cancel(reservationID)
}
