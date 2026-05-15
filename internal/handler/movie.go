package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"movie_reservation_system/internal/apperr"
	"movie_reservation_system/internal/service"
)

type MovieHandler struct {
	movieService *service.MovieService
}

func NewMovieHandler(movieService *service.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

type movieRequest struct {
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	PosterImageURL string `json:"poster_image_url"`
	Genre          string `json:"genre"`
}

func (h *MovieHandler) List(c *gin.Context) {
	movies, err := h.movieService.ListMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}
	c.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	movie, err := h.movieService.GetMovie(id)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movie"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) Create(c *gin.Context) {
	var req movieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie, err := h.movieService.CreateMovie(req.Title, req.Description, req.PosterImageURL, req.Genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create movie"})
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func (h *MovieHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req movieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie, err := h.movieService.UpdateMovie(id, req.Title, req.Description, req.PosterImageURL, req.Genre)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update movie"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.movieService.DeleteMovie(id); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "movie deleted"})
}

func parseID(c *gin.Context) (uint, error) {
	return parseIDFromParam(c, "id")
}

func parseIDFromParam(c *gin.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	return uint(id), err
}
