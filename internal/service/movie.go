package service

import (
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type MovieService struct {
	movieRepo *repository.MovieRepository
}

func NewMovieService(movieRepo *repository.MovieRepository) *MovieService {
	return &MovieService{movieRepo: movieRepo}
}

func (s *MovieService) ListMovies() ([]model.Movie, error) {
	return s.movieRepo.FindAll()
}

func (s *MovieService) GetMovie(id uint) (*model.Movie, error) {
	return s.movieRepo.FindByID(id)
}

func (s *MovieService) CreateMovie(title, description, posterURL, genre string) (*model.Movie, error) {
	movie := &model.Movie{
		Title:          title,
		Description:    description,
		PosterImageURL: posterURL,
		Genre:          genre,
	}
	if err := s.movieRepo.Create(movie); err != nil {
		return nil, err
	}
	return movie, nil
}

func (s *MovieService) UpdateMovie(id uint, title, description, posterURL, genre string) (*model.Movie, error) {
	movie, err := s.movieRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	movie.Title = title
	movie.Description = description
	movie.PosterImageURL = posterURL
	movie.Genre = genre

	if err := s.movieRepo.Update(movie); err != nil {
		return nil, err
	}
	return movie, nil
}

func (s *MovieService) DeleteMovie(id uint) error {
	if _, err := s.movieRepo.FindByID(id); err != nil {
		return err
	}
	return s.movieRepo.Delete(id)
}
