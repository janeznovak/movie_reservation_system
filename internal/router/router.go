package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"movie_reservation_system/internal/config"
	"movie_reservation_system/internal/handler"
	"movie_reservation_system/internal/middleware"
	"movie_reservation_system/internal/repository"
	"movie_reservation_system/internal/service"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	venueRepo := repository.NewVenueRepository(db)
	showtimeRepo := repository.NewShowtimeRepository(db)
	seatHoldRepo := repository.NewSeatHoldRepository(db)
	reservationRepo := repository.NewReservationRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	movieService := service.NewMovieService(movieRepo)
	venueService := service.NewVenueService(venueRepo)
	showtimeService := service.NewShowtimeService(showtimeRepo, movieRepo, venueRepo)
	seatHoldService := service.NewSeatHoldService(seatHoldRepo, showtimeRepo, venueRepo)
	reservationService := service.NewReservationService(reservationRepo)
	adminService := service.NewAdminService(reservationRepo, userRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	movieHandler := handler.NewMovieHandler(movieService)
	venueHandler := handler.NewVenueHandler(venueService)
	showtimeHandler := handler.NewShowtimeHandler(showtimeService)
	seatHoldHandler := handler.NewSeatHoldHandler(seatHoldService)
	reservationHandler := handler.NewReservationHandler(reservationService)
	adminHandler := handler.NewAdminHandler(adminService)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public auth routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Public movie routes
	movies := r.Group("/api/movies")
	{
		movies.GET("", movieHandler.List)
		movies.GET("/:id", movieHandler.Get)
		movies.GET("/:id/showtimes", showtimeHandler.ListByMovie)
	}

	// Public showtime routes
	r.GET("/api/showtimes", showtimeHandler.ListByDate)
	r.GET("/api/showtimes/:id/seats", seatHoldHandler.GetAvailableSeats)

	// Authenticated user routes
	user := r.Group("/api", middleware.AuthRequired(cfg.JWTSecret))
	{
		user.POST("/showtimes/:id/holds", seatHoldHandler.HoldSeats)
		user.DELETE("/holds/:id", seatHoldHandler.ReleaseHold)

		user.GET("/reservations", reservationHandler.List)
		user.POST("/reservations", reservationHandler.Create)
		user.DELETE("/reservations/:id", reservationHandler.Cancel)
	}

	// Admin-only routes
	admin := r.Group("/api/admin", middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
	{
		admin.POST("/movies", movieHandler.Create)
		admin.PUT("/movies/:id", movieHandler.Update)
		admin.DELETE("/movies/:id", movieHandler.Delete)

		admin.POST("/showtimes", showtimeHandler.Create)
		admin.PUT("/showtimes/:id", showtimeHandler.Update)
		admin.DELETE("/showtimes/:id", showtimeHandler.Delete)

		admin.GET("/venues", venueHandler.ListVenues)
		admin.POST("/venues", venueHandler.CreateVenue)
		admin.GET("/venues/:id/seats", venueHandler.ListSeats)
		admin.POST("/venues/:id/seats", venueHandler.AddSeat)
		admin.DELETE("/venues/:id/seats/:seat_id", venueHandler.RemoveSeat)

		admin.GET("/reservations", adminHandler.ListReservations)
		admin.GET("/reports", adminHandler.GetReport)
		admin.PUT("/users/:id/promote", adminHandler.PromoteUser)
	}

	return r
}
