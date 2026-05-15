package main

import (
	"log"

	"movie_reservation_system/internal/config"
	"movie_reservation_system/internal/database"
	"movie_reservation_system/internal/router"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg)
	database.RunMigrations(cfg)
	database.SeedAdmin(db)

	go database.CleanupExpiredHolds(db)

	r := router.Setup(db, cfg)
	log.Fatal(r.Run(":" + cfg.Port))
}
