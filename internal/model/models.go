package model

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type ReservationStatus string

const (
	StatusConfirmed ReservationStatus = "confirmed"
	StatusCancelled ReservationStatus = "cancelled"
)

type User struct {
	ID           uint           `gorm:"primarykey"  json:"id"`
	CreatedAt    time.Time      `                   json:"created_at"`
	UpdatedAt    time.Time      `                   json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"       json:"-"`
	Name         string         `gorm:"not null"                                 json:"name"`
	Email        string         `gorm:"uniqueIndex;not null"                     json:"email"`
	PasswordHash string         `gorm:"not null"                                 json:"-"`
	Role         Role           `gorm:"type:varchar(50);not null;default:'user'" json:"role"`
}

type Movie struct {
	ID             uint           `gorm:"primarykey"  json:"id"`
	CreatedAt      time.Time      `                   json:"created_at"`
	UpdatedAt      time.Time      `                   json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"       json:"-"`
	Title          string         `gorm:"not null"                json:"title"`
	Description    string         `                               json:"description"`
	PosterImageURL string         `gorm:"column:poster_image_url" json:"poster_image_url"`
	Genre          string         `                               json:"genre"`
}

type Venue struct {
	ID        uint           `gorm:"primarykey"  json:"id"`
	CreatedAt time.Time      `                   json:"created_at"`
	UpdatedAt time.Time      `                   json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"       json:"-"`
	Name      string         `gorm:"not null" json:"name"`
	Seats     []Seat         `                json:"seats,omitempty"`
}

type Seat struct {
	ID         uint           `gorm:"primarykey"  json:"id"`
	CreatedAt  time.Time      `                   json:"created_at"`
	UpdatedAt  time.Time      `                   json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"       json:"-"`
	VenueID    uint           `gorm:"not null;index" json:"venue_id"`
	Venue      Venue          `                      json:"-"`
	RowLabel   string         `gorm:"not null"       json:"row_label"`
	SeatNumber int            `gorm:"not null"       json:"seat_number"`
}

type Showtime struct {
	ID        uint           `gorm:"primarykey"  json:"id"`
	CreatedAt time.Time      `                   json:"created_at"`
	UpdatedAt time.Time      `                   json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"       json:"-"`
	MovieID   uint           `gorm:"not null;index" json:"movie_id"`
	Movie     Movie          `                      json:"movie,omitempty"`
	VenueID   uint           `gorm:"not null;index" json:"venue_id"`
	Venue     Venue          `                      json:"venue,omitempty"`
	StartTime time.Time      `gorm:"not null;index" json:"start_time"`
	EndTime   time.Time      `gorm:"not null"       json:"end_time"`
	Price     float64        `gorm:"not null"       json:"price"`
}

type Reservation struct {
	ID         uint              `gorm:"primarykey"  json:"id"`
	CreatedAt  time.Time         `                   json:"created_at"`
	UpdatedAt  time.Time         `                   json:"updated_at"`
	DeletedAt  gorm.DeletedAt    `gorm:"index"       json:"-"`
	UserID     uint              `gorm:"not null;index"                               json:"user_id"`
	User       *User             `                                                    json:"user,omitempty"`
	ShowtimeID uint              `gorm:"not null;index"                               json:"showtime_id"`
	Showtime   Showtime          `                                                    json:"showtime,omitempty"`
	Status     ReservationStatus `gorm:"type:varchar(50);not null;default:'confirmed'" json:"status"`
	TotalPrice float64           `gorm:"not null"                                     json:"total_price"`
	Seats      []Seat            `gorm:"many2many:reservation_seats;"                 json:"seats,omitempty"`
}

type ShowtimeReport struct {
	ShowtimeID     uint      `json:"showtime_id"`
	Movie          string    `json:"movie"`
	Venue          string    `json:"venue"`
	StartTime      time.Time `json:"start_time"`
	TotalSeats     int       `json:"total_seats"`
	SeatsSold      int       `json:"seats_sold"`
	SeatsAvailable int       `json:"seats_available"`
	Revenue        float64   `json:"revenue"`
}

type SeatHold struct {
	ID         uint           `gorm:"primarykey"  json:"id"`
	CreatedAt  time.Time      `                   json:"created_at"`
	UpdatedAt  time.Time      `                   json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"       json:"-"`
	SeatID     uint           `gorm:"not null;index" json:"seat_id"`
	Seat       Seat           `                      json:"seat,omitempty"`
	ShowtimeID uint           `gorm:"not null;index" json:"showtime_id"`
	Showtime   *Showtime      `                      json:"showtime,omitempty"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	User       User           `                      json:"-"`
	ExpiresAt  time.Time      `gorm:"not null;index" json:"expires_at"`
}
