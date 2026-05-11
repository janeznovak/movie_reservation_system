# Movie Reservation System

A backend REST API for a movie reservation service, built as a learning project.

## Tech Stack

- **Language**: Go
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL (via docker-compose)
- **Auth**: JWT
- **Migrations**: golang-migrate (versioned SQL files in `/migrations`)

## Project Structure

```
movie_reservation_system/
├── cmd/api/main.go          # Entry point
├── internal/
│   ├── config/              # Config loading from env
│   ├── handler/             # Gin HTTP handlers
│   ├── service/             # Business logic
│   ├── repository/          # GORM database access
│   ├── model/               # GORM model structs
│   ├── middleware/          # JWT auth middleware
│   └── router/              # Gin route registration
├── migrations/              # SQL migration files (up/down)
├── docker-compose.yml
├── .env
└── go.mod
```

Architecture is **layered**: handler → service → repository. Handlers parse HTTP, services contain business logic, repositories talk to the DB.

## Data Model

```
users
  id, name, email, password_hash, role (admin|user), created_at, updated_at

movies
  id, title, description, poster_image_url, genre (string), created_at, updated_at

venues
  id, name, created_at, updated_at

seats
  id, venue_id, row_label (e.g. "A"), seat_number (e.g. 1), created_at
  → admin manually adds seats to a venue

showtimes
  id, movie_id, venue_id, start_time, end_time, price, created_at, updated_at

reservations
  id, user_id, showtime_id, status (confirmed|cancelled), total_price, created_at, updated_at
  → total_price = showtime.price × number of seats, stored at booking time

reservation_seats
  id, reservation_id, seat_id

seat_holds
  id, seat_id, showtime_id, user_id, expires_at, created_at
  → temporary hold while user is browsing/paying (default: 10 minutes)
  → a seat is unavailable if in reservation_seats OR in seat_holds WHERE expires_at > now()
```

## API Endpoints

### Auth (public)
- `POST /api/auth/register`
- `POST /api/auth/login`

### Movies
- `GET /api/movies` — list all movies (public)
- `GET /api/movies/:id` — movie detail (public)
- `POST /api/movies` — create (admin)
- `PUT /api/movies/:id` — update (admin)
- `DELETE /api/movies/:id` — delete (admin)

### Showtimes
- `GET /api/showtimes?date=YYYY-MM-DD` — showtimes for a date (public)
- `GET /api/movies/:id/showtimes` — showtimes for a movie (public)
- `POST /api/showtimes` — create (admin)
- `PUT /api/showtimes/:id` — update (admin)
- `DELETE /api/showtimes/:id` — delete (admin)

### Venues & Seats (admin only)
- `GET /api/venues` — list venues
- `POST /api/venues` — create venue
- `GET /api/venues/:id/seats` — list seats for a venue
- `POST /api/venues/:id/seats` — add a seat to a venue
- `DELETE /api/venues/:id/seats/:seat_id` — remove a seat

### Seat Availability & Holds
- `GET /api/showtimes/:id/seats` — available seats for a showtime (public)
- `POST /api/showtimes/:id/holds` — hold selected seats for 10 minutes (user)
- `DELETE /api/holds/:id` — release a hold early (user)

### Reservations
- `GET /api/reservations` — my reservations (user)
- `POST /api/reservations` — convert a hold into a confirmed reservation (user)
- `DELETE /api/reservations/:id` — cancel reservation, upcoming only (user)

### Admin
- `GET /api/admin/reservations` — all reservations
- `GET /api/admin/reports` — capacity and revenue per showtime
- `PUT /api/admin/users/:id/promote` — promote user to admin

## Key Implementation Notes

### Overbooking prevention
Use a DB transaction with `SELECT FOR UPDATE` when placing a hold or confirming a reservation. Lock the relevant rows so two concurrent requests cannot claim the same seat.

### Seat holds
- Holding seats: INSERT into `seat_holds` (expires_at = now() + 10 min) inside a transaction with SELECT FOR UPDATE
- Confirming a reservation: verify the user's hold still exists AND expires_at > now(), then DELETE the hold and INSERT into reservations + reservation_seats
- If the hold expired before the user paid: return a 409 error — "your hold timed out, please select seats again"
- Background goroutine runs at startup, deletes expired holds every 5 minutes (cleanup only — expired rows are already ignored by queries)

### JWT
- Sign with a secret from env (`JWT_SECRET`)
- Include `user_id` and `role` in the claims
- Auth middleware validates token and injects claims into Gin context

### Roles
- `admin`: manage movies, venues, showtimes, seats; view all reservations and reports; promote users
- `user`: browse movies/showtimes, reserve seats, view/cancel own reservations

### Initial admin
Created via seed data (SQL or a Go seed script run at startup if no admin exists).

### Price
Static per showtime. `total_price` on a reservation is computed at booking time and stored — preserves historical accuracy if the showtime price changes later.

### Cancellation
Only reservations where `showtime.start_time > now()` can be cancelled.

## Development Setup

```bash
# Start the database
docker-compose up -d

# Run migrations
migrate -path migrations -database "postgres://..." up

# Run the server
go run cmd/api/main.go
```

Environment variables (`.env`):
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=movie_reservation
JWT_SECRET=your-secret-here
PORT=8080
```
