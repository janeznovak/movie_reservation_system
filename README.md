# Movie Reservation System

A backend REST API for a movie reservation service built in Go. This is a learning project focused on complex business logic — seat reservation, scheduling, data modelling, and concurrent access patterns.

## Tech Stack

| Tool | Purpose |
|------|---------|
| Go | Language |
| Gin | HTTP framework |
| GORM | ORM (database access) |
| PostgreSQL | Database |
| golang-migrate | Database migrations |
| JWT | Authentication |
| docker-compose | Local database setup |

---

## Phase 0 — Design & Planning

Before writing any code we designed the full system: data model, API endpoints, and key technical decisions.

### Data Model

```
users             → stores accounts with a role (admin or user)
movies            → title, description, poster URL, genre
venues            → a physical location (room, outdoor, etc.)
seats             → individual seats belonging to a venue, added manually by admins
showtimes         → a movie playing at a venue at a specific time, with a fixed price
reservations      → a confirmed booking by a user for a showtime
reservation_seats → which specific seats belong to a reservation (one reservation, many seats)
seat_holds        → temporary holds on seats while a user is browsing/paying
```

### Why one reservation with multiple seats?

Two approaches were considered:

- **Multiple reservations** (one per seat) — simple, but if a user books 3 seats together there's no link between them. Cancelling a "group" is awkward.
- **One reservation + reservation_seats table** — matches how real booking systems work. Easy to cancel a whole booking, easy to report revenue per transaction.

We chose the second approach.

### Why a `venues` table instead of `halls` or `rooms`?

The system should work for any kind of cinema — not just traditional rooms. A venue can be a room, a rooftop, a park, etc. Seats are defined per venue by an admin, rather than auto-generated.

### Concepts

#### SELECT FOR UPDATE

When two users try to book the same seat at the same time, a regular read-then-write is not safe:

```
User A reads: seat A3 is free
User B reads: seat A3 is free   ← both read simultaneously
User A books seat A3
User B books seat A3            ← overbooking!
```

`SELECT FOR UPDATE` locks the rows for the duration of a database transaction, forcing other requests to wait:

```
User A locks the showtime row
User B waits...
User A checks seats, books, commits → lock released
User B proceeds, checks seats → already taken → returns error to user
```

The lock only lasts milliseconds (one transaction). Regular reads like browsing available seats are plain `SELECT` statements and are never blocked.

#### Seat Holds (Soft Reservations)

Many ticketing systems let you "hold" seats for a few minutes while you complete payment, without committing. We implement this with a `seat_holds` table:

- User selects seats → hold created with `expires_at = now() + 10 minutes`
- Seat appears unavailable to others during that window
- User completes payment → hold is deleted, real reservation is created
- User abandons / timer runs out → hold is ignored automatically by queries (`WHERE expires_at > now()`)

A background goroutine cleans up expired hold rows every 5 minutes. This is purely housekeeping — correctness is guaranteed by the `expires_at` filter in queries, not by the cleanup job.

#### Goroutines (Background Jobs)

In Go, adding `go` before a function call runs it concurrently in the background:

```go
go cleanupExpiredHolds()  // runs in background, main continues immediately
```

Our cleanup job is a goroutine that loops forever — sleep 5 minutes, delete expired rows, repeat. It starts once when the server starts and runs for the lifetime of the app.

---

## Phase 1 — Project Scaffold

Set up the full project skeleton: folder structure, dependencies, database connection, migrations, and a running server with one health-check endpoint.

### What was built

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Runs PostgreSQL locally with one command |
| `.env` | All config in one place (ignored by git) |
| `internal/config/config.go` | Loads env vars into a typed struct |
| `internal/model/models.go` | All 6 GORM models + role/status constants |
| `internal/database/db.go` | DB connect, run migrations, seed admin, cleanup goroutine |
| `internal/router/router.go` | Gin setup with a `/health` endpoint |
| `cmd/api/main.go` | Entry point — wires everything together |
| `migrations/000001–000006` | SQL up/down files for every table |

### How to run

```bash
docker-compose up -d       # start PostgreSQL
go run cmd/api/main.go     # start the server (runs migrations + seeds admin automatically)
```

Default admin account created on first run: `admin@example.com` / `Admin1234!`

### Concepts

#### Database Migrations

Rather than letting the ORM create tables automatically, we use versioned SQL files in `migrations/`. Each change to the schema gets its own numbered file pair:

```
000001_create_users.up.sql    ← applies the change
000001_create_users.down.sql  ← reverses it
```

This gives a clear history of every schema change, lets you roll back safely, and reflects how real production systems are managed.

#### Indexes

An index is a separate data structure PostgreSQL maintains alongside a table so it can find rows quickly without scanning every row.

Without an index on `start_time`, a query like `WHERE start_time = '2026-05-12'` would read every row in the showtimes table. With an index, PostgreSQL jumps directly to the matching rows.

**The rule of thumb: if you frequently filter, sort, or join on a column — index it.**

Common signals to add an index:
- **Foreign keys** (`venue_id`, `user_id`, etc.) — you almost always join on these
- **Columns in WHERE clauses** — `email` for login lookups, `expires_at` for hold availability checks
- **Columns in ORDER BY** — if results are frequently sorted by a column

Signal to skip an index:
- Columns you never filter by (e.g. `description`, `poster_image_url`) — indexing them wastes space and slows down writes for no benefit

Note: primary keys (`id`) are **automatically indexed** by PostgreSQL — never create one manually for those.

**The tradeoff:** indexes speed up reads but slow down writes slightly (every insert/update/delete must also update the index) and use disk space. Index selectively.

#### Soft Deletes

GORM's `gorm.Model` adds a `deleted_at` column to every table. Instead of physically removing a row when you delete something, GORM sets `deleted_at` to the current timestamp and then automatically adds `WHERE deleted_at IS NULL` to every query — so deleted records are invisible but still in the database.

This is why every table has a `deleted_at` index: since that filter runs on every single query, without the index PostgreSQL would scan the whole table every time.

---

## Phase 2 — Authentication

Implemented user registration, login, JWT token generation, and auth middleware.

### What was built

| File | Purpose |
|------|---------|
| `internal/repository/user.go` | DB queries — find user by email, create user |
| `internal/service/auth.go` | Business logic — hash password, verify credentials, generate JWT |
| `internal/handler/auth.go` | HTTP handlers for register and login |
| `internal/middleware/auth.go` | JWT validation middleware + admin role check |
| `internal/router/router.go` | Wired up auth routes |

### Endpoints

```
POST /api/auth/register   → creates account, returns JWT
POST /api/auth/login      → verifies credentials, returns JWT
```

### Concepts

#### Layered architecture in practice

Each layer has one responsibility and only talks to the layer below it:

```
Handler  → parses HTTP request, validates input, returns HTTP response
Service  → contains business logic (is the email taken? is the password correct?)
Repository → talks to the database (find user, create user)
```

This separation means if you swap PostgreSQL for another database, only the repository changes. If you change the HTTP framework, only the handler changes.

#### Password hashing with bcrypt

Passwords are never stored in plain text. Before saving, the password is run through **bcrypt**, which is a one-way hashing function — you cannot reverse it to get the original password back.

When a user logs in, we run bcrypt on the password they provide and compare the result to the stored hash. If they match, the password is correct.

bcrypt also adds a random **salt** to each hash automatically, so two users with the same password will have different hashes in the database.

#### JWT (JSON Web Token)

After a successful login or register, the server returns a JWT token. This token is a signed string the client stores and sends with every future request in the `Authorization` header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

The token contains a **payload** (not encrypted, just encoded):
```json
{ "user_id": 2, "role": "user", "exp": 1778788129 }
```

The server signs the token with a secret key. On every request, the middleware verifies the signature — if anyone tampers with the payload, the signature won't match and the token is rejected.

This means the server never needs to look up a session in the database on every request — all the information it needs is inside the token itself.

#### Why we return the same error for wrong email and wrong password

```go
// Login returns the same error whether the email doesn't exist or password is wrong
return "", errors.New("invalid credentials")
```

If we returned different errors ("email not found" vs "wrong password"), an attacker could use the login endpoint to discover which emails have accounts. A generic message reveals nothing.

#### Middleware chaining

Admin-only routes use two middlewares in sequence:

```go
// AuthRequired validates the JWT and injects user_id + role into context
// AdminRequired reads the role from context and rejects non-admins
r.Group("/api/admin", middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
```

`AuthRequired` must run first — it sets the role. `AdminRequired` depends on that value. If `AuthRequired` rejects the request, Gin calls `c.Abort()` and `AdminRequired` never runs.

---

## Phase 3 — Movie Management

Implemented full CRUD for movies, with public read access and admin-only write access.

### What was built

| File | Purpose |
|------|---------|
| `internal/repository/movie.go` | DB queries — list, find, create, update, delete |
| `internal/service/movie.go` | Business logic — validates movie exists before update/delete |
| `internal/handler/movie.go` | HTTP handlers for all 5 movie endpoints |
| `internal/router/router.go` | Wired public routes + admin-protected routes |

### Endpoints

```
GET    /api/movies         → list all movies (public)
GET    /api/movies/:id     → single movie (public)
POST   /api/admin/movies   → create movie (admin)
PUT    /api/admin/movies/:id → update movie (admin)
DELETE /api/admin/movies/:id → delete movie (admin)
```

### Concepts

#### JSON struct tags

Go struct fields are exported (public) when they start with a capital letter — that's required for GORM to read/write them. But by default, `encoding/json` serializes them with the same capital-letter names (`ID`, `CreatedAt`), which looks wrong in an API response.

Struct tags let you control how a field is serialized without changing its name in Go code:

```go
Title string `gorm:"not null" json:"title"`
```

The `json:"title"` tag tells the JSON encoder to use `"title"` in the output instead of `"Title"`. You can also use:
- `json:"-"` — completely hide the field from JSON (used for `PasswordHash`, `DeletedAt`)
- `json:"seats,omitempty"` — omit the field if it's empty/nil

#### Why CRUD needs a "find before update/delete"

Before updating or deleting, the service first fetches the record:

```go
func (s *MovieService) DeleteMovie(id uint) error {
    _, err := s.movieRepo.FindByID(id)
    if err != nil {
        return err  // returns 404 if not found
    }
    return s.movieRepo.Delete(id)
}
```

Without this, deleting a non-existent ID would silently succeed (GORM's DELETE returns no error if 0 rows are affected). The explicit fetch ensures we return a proper 404 to the client.

#### HTTP status codes used

| Code | Meaning | When we use it |
|------|---------|----------------|
| 200 | OK | Successful GET, PUT, DELETE |
| 201 | Created | Successful POST (resource created) |
| 400 | Bad Request | Invalid input / missing required field |
| 401 | Unauthorized | Missing or invalid JWT token |
| 403 | Forbidden | Valid token but wrong role (user trying admin endpoint) |
| 404 | Not Found | Movie/resource doesn't exist |
| 409 | Conflict | Duplicate (e.g. email already in use) |
| 500 | Internal Server Error | Unexpected DB or server error |

---

## Error Handling — Option B (custom error types)

### Where status codes are decided

Status codes are chosen **only in the handler layer** — nowhere else. The flow is:

```
Repository  → executes SQL, returns raw GORM error (knows nothing about HTTP)
    ↓
Service     → translates GORM errors into our own errors (still knows nothing about HTTP)
    ↓
Handler     → checks our own errors, picks the status code, writes the response
```

The handler is the only layer that knows about HTTP, so it's the only layer that picks numbers like 404 or 500.

### Why three layers?

Each layer has one job. Things change independently:

| Layer | Its job | What would change independently |
|-------|---------|----------------------------------|
| Repository | Talk to the DB | Swapping PostgreSQL for another DB |
| Service | Business logic | Changing rules ("only admins can delete") |
| Handler | HTTP in/out | Swapping Gin for another framework |

Without this separation, a database change forces you to touch HTTP code. With it, each change stays in its own box.

### Why not check `gorm.ErrRecordNotFound` directly in the handler?

The handler's job is HTTP — it shouldn't need to know which database library the service uses. If we import `gorm` in the handler just to check an error type, we've leaked a database concern into the HTTP layer.

The fix is a small `apperr` package with our own sentinel errors:

```go
// internal/apperr/errors.go
var (
    ErrNotFound = errors.New("not found")
    ErrConflict = errors.New("already exists")
)
```

The **service** translates GORM errors into these before returning them:

```go
// service/movie.go — GORM error stays inside the service
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, apperr.ErrNotFound
}
```

The **handler** only knows about `apperr` — no GORM import needed:

```go
// handler/movie.go — clean, no database knowledge here
if errors.Is(err, apperr.ErrNotFound) {
    c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
}
```

### Where should the GORM → apperr translation live?

The standard practice is to translate in the **repository layer**. The repository is the only layer that knows about GORM — service and handler layers should have no knowledge of which database library is being used.

```
Handler     → knows HTTP (status codes, JSON)
Service     → knows business logic
Repository  → knows GORM/database
```

Each layer should only deal with errors from its own domain. GORM errors are a database concern, so they should never leave the repository:

```go
// repository — catches GORM error, translates before returning
func (r *MovieRepository) FindByID(id uint) (*model.Movie, error) {
    var movie model.Movie
    err := r.db.First(&movie, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperr.ErrNotFound  // GORM error stays here
    }
    return &movie, err
}

// service — knows nothing about GORM
func (s *MovieService) GetMovie(id uint) (*model.Movie, error) {
    return s.movieRepo.FindByID(id)  // already returns apperr if not found
}

// handler — knows nothing about GORM
if errors.Is(err, apperr.ErrNotFound) {
    c.JSON(http.StatusNotFound, ...)
}
```

If you ever swap GORM for another library, only the repository changes — service and handler code stays untouched.

Our implementation follows this standard — GORM error translation happens in the repository, and services and handlers only ever see `apperr` errors.

### How GORM error returns work

A GORM call like `db.Delete(...)` always executes the SQL immediately. It returns a `*gorm.DB` result object, which has an `.Error` field:

```go
return r.db.Delete(&model.Movie{}, id).Error
// DELETE runs → success → .Error is nil  → caller sees no error
// DELETE runs → failure → .Error is set  → caller sees the error
```

The return does not control whether the SQL runs — it only passes back whether it succeeded. `nil` means success, non-nil means something went wrong.

---

## Phase 4 — Venues & Seats

Implemented admin-only venue management and manual seat definition per venue.

### What was built

| File | Purpose |
|------|---------|
| `internal/repository/venue.go` | DB queries for venues and seats, including constraint violation detection |
| `internal/service/venue.go` | Business logic — validates venue exists before seat operations, verifies seat belongs to venue on delete |
| `internal/handler/venue.go` | HTTP handlers for all 5 venue/seat endpoints |

### Endpoints

```
GET    /api/admin/venues                      → list all venues
POST   /api/admin/venues                      → create venue
GET    /api/admin/venues/:id/seats            → list seats for a venue
POST   /api/admin/venues/:id/seats            → add a seat to a venue
DELETE /api/admin/venues/:id/seats/:seat_id   → remove a seat
```

### Concepts

#### PostgreSQL unique constraint violations

When an admin tries to add the same seat twice (e.g. row A, number 1 already exists in venue 1), the database rejects the insert with a unique constraint violation. This is a different kind of error than "record not found" — it comes back as a PostgreSQL error code `23505`.

GORM doesn't have a built-in sentinel for this like it does for `ErrRecordNotFound`, so we detect it by inspecting the underlying PostgreSQL error directly:

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    return apperr.ErrConflict
}
```

`errors.As` unwraps the error chain to find a `*pgconn.PgError` — the raw PostgreSQL error type. If the code is `23505` (PostgreSQL's code for unique violation), we translate it to our own `ErrConflict` domain error.

#### Ownership check on delete

When removing a seat, we verify it actually belongs to the venue in the URL — not just that it exists:

```go
seat, err := s.venueRepo.FindSeatByID(seatID)
if seat.VenueID != venueID {
    return apperr.ErrNotFound
}
```

Without this check, someone could delete a seat from venue 2 by calling `DELETE /venues/1/seats/5` — even though seat 5 belongs to venue 2. The ownership check ensures the URL is consistent with the data.

---

## Phase 5 — Showtimes

Implemented showtime management for admins and public browsing by date or movie.

### What was built

| File | Purpose |
|------|---------|
| `internal/repository/showtime.go` | DB queries — find by ID, by date range, by movie; CRUD |
| `internal/service/showtime.go` | Business logic — validates movie/venue exist, validates times |
| `internal/handler/showtime.go` | HTTP handlers for all showtime endpoints |

### Endpoints

```
GET /api/showtimes?date=YYYY-MM-DD    → all showtimes on a date (public)
GET /api/movies/:id/showtimes         → showtimes for a specific movie (public)
POST   /api/admin/showtimes           → create showtime (admin)
PUT    /api/admin/showtimes/:id       → update showtime (admin)
DELETE /api/admin/showtimes/:id       → delete showtime (admin)
```

### Concepts

#### Preloading related records (GORM)

When fetching a showtime, the response should include the movie and venue details — not just their IDs. GORM's `Preload` handles this:

```go
r.db.Preload("Movie").Preload("Venue").First(&showtime, id)
```

This runs two extra SELECT queries behind the scenes — one for the movie, one for the venue — and attaches the results to the struct. Without `Preload`, the `Movie` and `Venue` fields in the struct would be empty even though `movie_id` and `venue_id` are set.

#### Date range query

Browsing showtimes by date uses a range rather than an exact timestamp match, since showtimes could start at any time during the day:

```go
start := beginning of the requested day (00:00:00)
end   := start + 24 hours

WHERE start_time >= start AND start_time < end
```

#### Input validation in the service

The showtime service validates two things before writing to the DB:

1. **References exist** — checks that the provided `movie_id` and `venue_id` actually exist, returning a clear error if not
2. **Time ordering** — rejects a showtime where `end_time` is not after `start_time`

This validation lives in the service (not the handler) because it's business logic, not just HTTP parsing.

---

## Phase 6 — Seat Availability & Holds

Implemented public seat browsing and authenticated seat holds with overbooking prevention using `SELECT FOR UPDATE`.

### What was built

| File | Purpose |
|------|---------|
| `internal/repository/seat_hold.go` | Availability query, transactional hold creation with SELECT FOR UPDATE, delete |
| `internal/service/seat_hold.go` | Validates showtime exists, validates seats belong to the venue, deduplicates seat IDs |
| `internal/handler/seat_hold.go` | HTTP handlers for the three endpoints |

### Endpoints

```
GET    /api/showtimes/:id/seats      → available seats for a showtime (public)
POST   /api/showtimes/:id/holds      → hold selected seats for 10 minutes (user, requires JWT)
DELETE /api/holds/:id                → release a hold early (user, requires JWT)
```

### Concepts

#### How seat availability is determined

A seat is unavailable for a given showtime if either:
- It appears in `reservation_seats` joined to a confirmed reservation for that showtime, OR
- It appears in `seat_holds` where `expires_at > now()` for that showtime

The availability query excludes both:

```sql
SELECT * FROM seats
WHERE venue_id = ?
  AND id NOT IN (
    SELECT seat_id FROM reservation_seats
    JOIN reservations ON reservations.id = reservation_seats.reservation_id
    WHERE reservations.showtime_id = ? AND reservations.status = 'confirmed'
  )
  AND id NOT IN (
    SELECT seat_id FROM seat_holds
    WHERE showtime_id = ? AND expires_at > now() AND deleted_at IS NULL
  )
ORDER BY row_label, seat_number
```

GORM adds `AND seats.deleted_at IS NULL` automatically. The subquery conditions must be written manually since they're raw SQL inside a `.Where()` string.

#### SELECT FOR UPDATE — preventing the race condition

Without a lock, two users can hold the same seat simultaneously:

```
User A: SELECT — seat 3 is free  ✓
User B: SELECT — seat 3 is free  ✓   (reads at the same moment)
User A: INSERT hold for seat 3
User B: INSERT hold for seat 3   ← overbooking!
```

With `SELECT FOR UPDATE`, the transaction acquires a row-level lock on the seat rows before checking:

```go
tx.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id IN ?", seatIDs).
    Find(&seats)
```

Now only one request can hold the lock at a time:

```
User A: locks seat 3, checks availability → free → inserts hold → commits → releases lock
User B: waits for lock → gets lock → checks availability → already held → returns 409
```

The lock lasts only for the duration of the transaction — milliseconds. Browsing (plain SELECT) is never blocked.

#### Why we return 404 when releasing another user's hold

When a user tries to release a hold they don't own:

```go
if hold.UserID != userID {
    return apperr.ErrNotFound
}
```

Returning 403 Forbidden would confirm to the caller that the hold exists. Returning 404 reveals nothing — the hold is invisible to other users, as if it doesn't exist. This is the same principle used for any resource that shouldn't be visible across accounts.

#### omitempty and pointer types

Go's `encoding/json` package only respects `omitempty` for basic types (string, int, bool) and nil pointers. It does NOT omit a zero-value struct:

```go
// This will still appear in JSON as an empty object even if never populated:
Showtime Showtime `json:"showtime,omitempty"`

// This will be omitted when nil — pointer lets omitempty work:
Showtime *Showtime `json:"showtime,omitempty"`
```

The `SeatHold` model uses `*Showtime` because the hold response doesn't preload the showtime (the client already knows which showtime they booked). Using a pointer means the field disappears from the JSON entirely when not populated.

---

## Phase 7 — Reservations

Implemented the full booking flow: converting holds into a confirmed reservation, listing a user's reservations, and cancellation.

### What was built

| File | Purpose |
|------|---------|
| `internal/repository/reservation.go` | Transactional hold-to-reservation conversion, find by user, cancel |
| `internal/service/reservation.go` | Validates ownership, cancellation eligibility, deduplicates hold IDs |
| `internal/handler/reservation.go` | HTTP handlers for the three endpoints |

### Endpoints

```
POST   /api/reservations      → confirm holds into a reservation (user, requires JWT)
GET    /api/reservations      → list my reservations (user, requires JWT)
DELETE /api/reservations/:id  → cancel an upcoming reservation (user, requires JWT)
```

### Concepts

#### Converting holds into a reservation

The request body is just the hold IDs the user wants to confirm:

```json
{ "hold_ids": [3, 4] }
```

Inside a transaction, the repository:
1. Locks the hold rows with `SELECT FOR UPDATE`
2. Validates all holds belong to the user and the same showtime
3. Validates none have expired (`expires_at > now()`)
4. Gets the showtime price, calculates `total_price = price × count`
5. Creates the reservation row
6. Inserts the seats into the `reservation_seats` join table via GORM's `Association("Seats").Append`
7. Soft-deletes the holds — they are no longer needed

If a hold expired between the user browsing and paying, the server returns 409 with the message "your hold timed out, please select seats again".

#### Why total_price is stored at booking time

The showtime price is copied into `total_price` on the reservation at the moment of booking:

```go
TotalPrice: showtime.Price * float64(len(holds)),
```

If an admin later changes the showtime price, existing reservations keep the price the user actually paid. Historical accuracy is preserved because the number is stored, not computed on the fly.

#### Cancellation rules

Two checks happen before cancelling:

1. **Ownership** — `reservation.UserID != userID` → 404 (same pattern as holds: don't reveal other users' data)
2. **Time** — `showtime.StartTime` must be in the future; you cannot cancel after the film has started

After cancellation the status changes to `cancelled`. The seats in `reservation_seats` are kept — they're historical record. The availability query already handles this by only excluding seats linked to `status = 'confirmed'` reservations, so cancelled reservation seats automatically become bookable again.

#### Nested preloads

The reservation response includes the full showtime (with movie and venue nested inside). GORM's preload accepts a callback function to chain further preloads on the relation:

```go
db.Preload("Showtime", func(db *gorm.DB) *gorm.DB {
    return db.Preload("Movie").Preload("Venue")
}).Preload("Seats").First(&reservation, id)
```

This runs three extra SELECT queries behind the scenes (showtime, movie, venue) and assembles them into the struct automatically.

#### New sentinel error: ErrExpired

The `apperr` package now has three sentinel errors:

```go
var (
    ErrNotFound = errors.New("not found")   // resource doesn't exist
    ErrConflict = errors.New("already exists")  // duplicate / seat taken
    ErrExpired  = errors.New("expired")     // hold timed out before confirmation
)
```

Each maps to a specific HTTP status code in the handler: 404, 409, and 409 respectively. Using named sentinel errors (rather than plain string messages) means the handler can check with `errors.Is` instead of comparing strings — which is more robust and doesn't break if the message wording changes.

---

## Phase 8 — Admin Features

Implemented the three remaining admin endpoints: viewing all reservations, a per-showtime report, and promoting a user to admin.

### What was built

| File | Purpose |
|------|---------|
| `internal/model/models.go` | Added `ShowtimeReport` result struct |
| `internal/repository/user.go` | Added `SetRole` — updates role, detects not-found via `RowsAffected` |
| `internal/repository/reservation.go` | Added `FindAll` (with user preloaded) and `GetReport` (raw SQL) |
| `internal/service/admin.go` | `AdminService` — thin layer wiring the three operations |
| `internal/handler/admin.go` | `AdminHandler` — three HTTP handlers |

### Endpoints

```
GET /api/admin/reservations    → all reservations with user, showtime, seats (admin)
GET /api/admin/reports         → capacity and revenue per showtime (admin)
PUT /api/admin/users/:id/promote → promote a user to admin (admin)
```

### Concepts

#### The report query

The report uses raw SQL because it aggregates across multiple tables in a way that doesn't map cleanly to GORM's query builder. Each row in the result is one showtime:

```sql
SELECT
    st.id AS showtime_id,
    m.title AS movie,
    v.name AS venue,
    st.start_time,
    (SELECT COUNT(*) FROM seats WHERE venue_id = st.venue_id ...)      AS total_seats,
    (SELECT COUNT(DISTINCT rs.seat_id) FROM reservation_seats rs
     JOIN reservations r ... WHERE r.showtime_id = st.id ...)          AS seats_sold,
    total_seats - seats_sold                                            AS seats_available,
    (SELECT COALESCE(SUM(total_price), 0) FROM reservations r ...)     AS revenue
FROM showtimes st
JOIN movies m ... JOIN venues v ...
```

`COALESCE(SUM(...), 0)` handles showtimes with no reservations — `SUM` of zero rows returns `NULL`, and `COALESCE` converts it to `0`.

GORM's `db.Raw(...).Scan(&results)` runs the raw SQL and maps column names to struct fields automatically (it matches `seats_sold` → `SeatsSold`, `start_time` → `StartTime`, etc.).

#### Detecting not-found without a prior lookup

For `PromoteUser`, rather than fetching the user first just to check it exists, the repository does the update and checks how many rows were affected:

```go
result := r.db.Model(&model.User{}).Where("id = ?", id).Update("role", role)
if result.RowsAffected == 0 {
    return apperr.ErrNotFound
}
```

If the user doesn't exist, the UPDATE touches 0 rows. One round trip to the database instead of two. This pattern applies whenever the update itself is the only thing you need — you don't need to read the record first.

#### Why the old JWT still shows the old role after promotion

JWT tokens are signed and self-contained — the server doesn't store them. Once issued, a token can't be changed. Promoting a user updates the database, but existing tokens are unaffected until they expire or the user logs in again and gets a fresh token.

This is a known trade-off of stateless JWT auth. Alternatives (like keeping a token blacklist or using short expiry times) add complexity. For this project, the 24-hour expiry and the expectation that the user logs in after promotion is acceptable.

---

## Project Complete

All 26 endpoints implemented across 8 phases:

| Phase | What | Endpoints |
|-------|------|-----------|
| 1 | Scaffold | `/health` |
| 2 | Auth | register, login |
| 3 | Movies | CRUD + list |
| 4 | Venues & Seats | venue CRUD, seat add/remove |
| 5 | Showtimes | CRUD + browse by date/movie |
| 6 | Seat Holds | availability, hold, release |
| 7 | Reservations | confirm, list mine, cancel |
| 8 | Admin | all reservations, reports, promote |
