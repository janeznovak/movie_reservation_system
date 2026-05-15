package service

import (
	"errors"
	"time"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type ShowtimeService struct {
	showtimeRepo *repository.ShowtimeRepository
	movieRepo    *repository.MovieRepository
	venueRepo    *repository.VenueRepository
}

func NewShowtimeService(
	showtimeRepo *repository.ShowtimeRepository,
	movieRepo *repository.MovieRepository,
	venueRepo *repository.VenueRepository,
) *ShowtimeService {
	return &ShowtimeService{
		showtimeRepo: showtimeRepo,
		movieRepo:    movieRepo,
		venueRepo:    venueRepo,
	}
}

func (s *ShowtimeService) GetShowtime(id uint) (*model.Showtime, error) {
	return s.showtimeRepo.FindByID(id)
}

func (s *ShowtimeService) ListByDate(date time.Time) ([]model.Showtime, error) {
	return s.showtimeRepo.FindByDate(date)
}

func (s *ShowtimeService) ListByMovie(movieID uint) ([]model.Showtime, error) {
	if _, err := s.movieRepo.FindByID(movieID); err != nil {
		return nil, err
	}
	return s.showtimeRepo.FindByMovieID(movieID)
}

func (s *ShowtimeService) CreateShowtime(movieID, venueID uint, startTime, endTime time.Time, price float64) (*model.Showtime, error) {
	if err := s.validateReferences(movieID, venueID); err != nil {
		return nil, err
	}
	if err := validateTimes(startTime, endTime); err != nil {
		return nil, err
	}

	showtime := &model.Showtime{
		MovieID:   movieID,
		VenueID:   venueID,
		StartTime: startTime,
		EndTime:   endTime,
		Price:     price,
	}
	if err := s.showtimeRepo.Create(showtime); err != nil {
		return nil, err
	}
	return s.showtimeRepo.FindByID(showtime.ID)
}

func (s *ShowtimeService) UpdateShowtime(id, movieID, venueID uint, startTime, endTime time.Time, price float64) (*model.Showtime, error) {
	showtime, err := s.showtimeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.validateReferences(movieID, venueID); err != nil {
		return nil, err
	}
	if err := validateTimes(startTime, endTime); err != nil {
		return nil, err
	}

	showtime.MovieID = movieID
	showtime.VenueID = venueID
	showtime.StartTime = startTime
	showtime.EndTime = endTime
	showtime.Price = price

	if err := s.showtimeRepo.Update(showtime); err != nil {
		return nil, err
	}
	return s.showtimeRepo.FindByID(showtime.ID)
}

func (s *ShowtimeService) DeleteShowtime(id uint) error {
	if _, err := s.showtimeRepo.FindByID(id); err != nil {
		return err
	}
	return s.showtimeRepo.Delete(id)
}

func (s *ShowtimeService) validateReferences(movieID, venueID uint) error {
	if _, err := s.movieRepo.FindByID(movieID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return errors.New("movie not found")
		}
		return err
	}
	if _, err := s.venueRepo.FindByID(venueID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return errors.New("venue not found")
		}
		return err
	}
	return nil
}

func validateTimes(start, end time.Time) error {
	if !end.After(start) {
		return errors.New("end_time must be after start_time")
	}
	return nil
}
