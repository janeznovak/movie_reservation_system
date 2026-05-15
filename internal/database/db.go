package database

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"movie_reservation_system/internal/config"
	"movie_reservation_system/internal/model"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	log.Println("database connected")
	return db
}

func RunMigrations(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatal("failed to init migrator: ", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrations: ", err)
	}

	log.Println("migrations applied")
}

func SeedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Admin1234!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash admin password: ", err)
	}

	admin := model.User{
		Name:         "Admin",
		Email:        "admin@example.com",
		PasswordHash: string(hash),
		Role:         model.RoleAdmin,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatal("failed to seed admin: ", err)
	}

	log.Println("admin seeded — email: admin@example.com  password: Admin1234!")
}

// CleanupExpiredHolds runs forever in a goroutine, deleting expired seat holds every 5 minutes.
// Unscoped() bypasses GORM's soft-delete so rows are physically removed.
func CleanupExpiredHolds(db *gorm.DB) {
	for {
		time.Sleep(5 * time.Minute)
		result := db.Unscoped().Where("expires_at < ?", time.Now()).Delete(&model.SeatHold{})
		if result.Error != nil {
			log.Println("hold cleanup error:", result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("cleaned up %d expired seat holds\n", result.RowsAffected)
		}
	}
}
