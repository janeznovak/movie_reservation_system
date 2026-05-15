package service

import (
	"errors"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type SeatHoldService struct {
	seatHoldRepo *repository.SeatHoldRepository
	showtimeRepo *repository.ShowtimeRepository
	venueRepo    *repository.VenueRepository
}

func NewSeatHoldService(
	seatHoldRepo *repository.SeatHoldRepository,
	showtimeRepo *repository.ShowtimeRepository,
	venueRepo *repository.VenueRepository,
) *SeatHoldService {
	return &SeatHoldService{
		seatHoldRepo: seatHoldRepo,
		showtimeRepo: showtimeRepo,
		venueRepo:    venueRepo,
	}
}

func (s *SeatHoldService) GetAvailableSeats(showtimeID uint) ([]model.Seat, error) {
	showtime, err := s.showtimeRepo.FindByID(showtimeID)
	if err != nil {
		return nil, err
	}
	return s.seatHoldRepo.FindAvailableSeats(showtime.VenueID, showtimeID)
}

func (s *SeatHoldService) HoldSeats(showtimeID, userID uint, seatIDs []uint) ([]model.SeatHold, error) {
	if len(seatIDs) == 0 {
		return nil, errors.New("at least one seat is required")
	}

	seen := make(map[uint]bool)
	unique := make([]uint, 0, len(seatIDs))
	for _, id := range seatIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	seatIDs = unique

	showtime, err := s.showtimeRepo.FindByID(showtimeID)
	if err != nil {
		return nil, err
	}

	venueSeats, err := s.venueRepo.FindSeatsByVenueID(showtime.VenueID)
	if err != nil {
		return nil, err
	}
	seatsInVenue := make(map[uint]bool, len(venueSeats))
	for _, seat := range venueSeats {
		seatsInVenue[seat.ID] = true
	}
	for _, seatID := range seatIDs {
		if !seatsInVenue[seatID] {
			return nil, apperr.ErrNotFound
		}
	}

	return s.seatHoldRepo.CreateHolds(showtimeID, userID, seatIDs)
}

func (s *SeatHoldService) ReleaseHold(holdID, userID uint) error {
	hold, err := s.seatHoldRepo.FindByID(holdID)
	if err != nil {
		return err
	}
	if hold.UserID != userID {
		return apperr.ErrNotFound
	}
	return s.seatHoldRepo.Delete(holdID)
}
