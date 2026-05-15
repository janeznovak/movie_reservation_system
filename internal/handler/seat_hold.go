package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/service"
)

type SeatHoldHandler struct {
	seatHoldService *service.SeatHoldService
}

func NewSeatHoldHandler(seatHoldService *service.SeatHoldService) *SeatHoldHandler {
	return &SeatHoldHandler{seatHoldService: seatHoldService}
}

type holdRequest struct {
	SeatIDs []uint `json:"seat_ids" binding:"required,min=1"`
}

func (h *SeatHoldHandler) GetAvailableSeats(c *gin.Context) {
	showtimeID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime id"})
		return
	}

	seats, err := h.seatHoldService.GetAvailableSeats(showtimeID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch seats"})
		return
	}

	c.JSON(http.StatusOK, seats)
}

func (h *SeatHoldHandler) HoldSeats(c *gin.Context) {
	showtimeID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime id"})
		return
	}

	userID := c.GetUint("user_id")

	var req holdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	holds, err := h.seatHoldService.HoldSeats(showtimeID, userID, req.SeatIDs)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "showtime or seat not found"})
			return
		}
		if errors.Is(err, apperr.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "one or more seats are already taken"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, holds)
}

func (h *SeatHoldHandler) ReleaseHold(c *gin.Context) {
	holdID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hold id"})
		return
	}

	userID := c.GetUint("user_id")

	if err := h.seatHoldService.ReleaseHold(holdID, userID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "hold not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release hold"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "hold released"})
}
