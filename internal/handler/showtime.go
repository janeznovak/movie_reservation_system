package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/service"
)

type ShowtimeHandler struct {
	showtimeService *service.ShowtimeService
}

func NewShowtimeHandler(showtimeService *service.ShowtimeService) *ShowtimeHandler {
	return &ShowtimeHandler{showtimeService: showtimeService}
}

type showtimeRequest struct {
	MovieID   uint    `json:"movie_id" binding:"required"`
	VenueID   uint    `json:"venue_id" binding:"required"`
	StartTime string  `json:"start_time" binding:"required"`
	EndTime   string  `json:"end_time" binding:"required"`
	Price     float64 `json:"price" binding:"required,gt=0"`
}

func (h *ShowtimeHandler) ListByDate(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query parameter is required (YYYY-MM-DD)"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	showtimes, err := h.showtimeService.ListByDate(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch showtimes"})
		return
	}

	c.JSON(http.StatusOK, showtimes)
}

func (h *ShowtimeHandler) ListByMovie(c *gin.Context) {
	movieID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	showtimes, err := h.showtimeService.ListByMovie(movieID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch showtimes"})
		return
	}

	c.JSON(http.StatusOK, showtimes)
}

func (h *ShowtimeHandler) Create(c *gin.Context) {
	var req showtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, endTime, err := parseTimes(req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showtime, err := h.showtimeService.CreateShowtime(req.MovieID, req.VenueID, startTime, endTime, req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, showtime)
}

func (h *ShowtimeHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req showtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, endTime, err := parseTimes(req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showtime, err := h.showtimeService.UpdateShowtime(id, req.MovieID, req.VenueID, startTime, endTime, req.Price)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, showtime)
}

func (h *ShowtimeHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.showtimeService.DeleteShowtime(id); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete showtime"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "showtime deleted"})
}

func parseTimes(startStr, endStr string) (time.Time, time.Time, error) {
	const layout = "2006-01-02T15:04:05Z07:00"
	start, err := time.Parse(layout, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid start_time format, use RFC3339 (e.g. 2026-05-14T20:00:00+02:00)")
	}
	end, err := time.Parse(layout, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("invalid end_time format, use RFC3339 (e.g. 2026-05-14T22:00:00+02:00)")
	}
	return start, end, nil
}
