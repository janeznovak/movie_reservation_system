package service

import (
	"movie_reservation_system/internal/model"
	"movie_reservation_system/internal/repository"
)

type AdminService struct {
	reservationRepo *repository.ReservationRepository
	userRepo        *repository.UserRepository
}

func NewAdminService(
	reservationRepo *repository.ReservationRepository,
	userRepo *repository.UserRepository,
) *AdminService {
	return &AdminService{
		reservationRepo: reservationRepo,
		userRepo:        userRepo,
	}
}

func (s *AdminService) ListAllReservations() ([]model.Reservation, error) {
	return s.reservationRepo.FindAll()
}

func (s *AdminService) GetReport() ([]model.ShowtimeReport, error) {
	return s.reservationRepo.GetReport()
}

func (s *AdminService) PromoteUser(userID uint) error {
	return s.userRepo.SetRole(userID, model.RoleAdmin)
}
