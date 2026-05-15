package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/service"
)

type VenueHandler struct {
	venueService *service.VenueService
}

func NewVenueHandler(venueService *service.VenueService) *VenueHandler {
	return &VenueHandler{venueService: venueService}
}

type createVenueRequest struct {
	Name string `json:"name" binding:"required"`
}

type addSeatRequest struct {
	RowLabel   string `json:"row_label" binding:"required"`
	SeatNumber int    `json:"seat_number" binding:"required"`
}

func (h *VenueHandler) ListVenues(c *gin.Context) {
	venues, err := h.venueService.ListVenues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch venues"})
		return
	}
	c.JSON(http.StatusOK, venues)
}

func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var req createVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	venue, err := h.venueService.CreateVenue(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create venue"})
		return
	}

	c.JSON(http.StatusCreated, venue)
}

func (h *VenueHandler) ListSeats(c *gin.Context) {
	venueID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid venue id"})
		return
	}

	seats, err := h.venueService.ListSeats(venueID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "venue not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch seats"})
		return
	}

	c.JSON(http.StatusOK, seats)
}

func (h *VenueHandler) AddSeat(c *gin.Context) {
	venueID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid venue id"})
		return
	}

	var req addSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	seat, err := h.venueService.AddSeat(venueID, req.RowLabel, req.SeatNumber)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "venue not found"})
			return
		}
		if errors.Is(err, apperr.ErrConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "seat already exists in this venue"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add seat"})
		return
	}

	c.JSON(http.StatusCreated, seat)
}

func (h *VenueHandler) RemoveSeat(c *gin.Context) {
	venueID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid venue id"})
		return
	}

	seatID, err := parseSeatID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seat id"})
		return
	}

	if err := h.venueService.RemoveSeat(venueID, seatID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "seat not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove seat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "seat removed"})
}

func parseSeatID(c *gin.Context) (uint, error) {
	return parseIDFromParam(c, "seat_id")
}
