package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/service"
)

type ReservationHandler struct {
	reservationService *service.ReservationService
}

func NewReservationHandler(reservationService *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{reservationService: reservationService}
}

type reservationRequest struct {
	HoldIDs []uint `json:"hold_ids" binding:"required,min=1"`
}

func (h *ReservationHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req reservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reservation, err := h.reservationService.CreateReservation(userID, req.HoldIDs)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hold not found"})
			return
		}
		if errors.Is(err, apperr.ErrExpired) {
			c.JSON(http.StatusConflict, gin.H{"error": "your hold timed out, please select seats again"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

func (h *ReservationHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")

	reservations, err := h.reservationService.ListReservations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reservations"})
		return
	}

	c.JSON(http.StatusOK, reservations)
}

func (h *ReservationHandler) Cancel(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID := c.GetUint("user_id")

	if err := h.reservationService.CancelReservation(id, userID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reservation cancelled"})
}
