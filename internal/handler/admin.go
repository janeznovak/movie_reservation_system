package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/service"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) ListReservations(c *gin.Context) {
	reservations, err := h.adminService.ListAllReservations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reservations"})
		return
	}
	c.JSON(http.StatusOK, reservations)
}

func (h *AdminHandler) GetReport(c *gin.Context) {
	report, err := h.adminService.GetReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *AdminHandler) PromoteUser(c *gin.Context) {
	userID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.adminService.PromoteUser(userID); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to promote user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user promoted to admin"})
}
