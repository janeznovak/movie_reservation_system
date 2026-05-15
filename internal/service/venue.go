package service

import (
	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type VenueService struct {
	venueRepo *repository.VenueRepository
}

func NewVenueService(venueRepo *repository.VenueRepository) *VenueService {
	return &VenueService{venueRepo: venueRepo}
}

func (s *VenueService) ListVenues() ([]model.Venue, error) {
	return s.venueRepo.FindAll()
}

func (s *VenueService) CreateVenue(name string) (*model.Venue, error) {
	venue := &model.Venue{Name: name}
	if err := s.venueRepo.Create(venue); err != nil {
		return nil, err
	}
	return venue, nil
}

func (s *VenueService) ListSeats(venueID uint) ([]model.Seat, error) {
	if _, err := s.venueRepo.FindByID(venueID); err != nil {
		return nil, err
	}
	return s.venueRepo.FindSeatsByVenueID(venueID)
}

func (s *VenueService) AddSeat(venueID uint, rowLabel string, seatNumber int) (*model.Seat, error) {
	if _, err := s.venueRepo.FindByID(venueID); err != nil {
		return nil, err
	}

	seat := &model.Seat{
		VenueID:    venueID,
		RowLabel:   rowLabel,
		SeatNumber: seatNumber,
	}
	if err := s.venueRepo.CreateSeat(seat); err != nil {
		return nil, err
	}
	return seat, nil
}

func (s *VenueService) RemoveSeat(venueID, seatID uint) error {
	seat, err := s.venueRepo.FindSeatByID(seatID)
	if err != nil {
		return err
	}
	// Ensure the seat actually belongs to the specified venue
	if seat.VenueID != venueID {
		return apperr.ErrNotFound
	}
	return s.venueRepo.DeleteSeat(seatID)
}
