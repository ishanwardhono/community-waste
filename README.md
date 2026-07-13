# Community Waste Collection API

REST API for a community waste collection system. Residents register their
households, request waste pickups (organic, plastic, paper, electronic), and
pay a fee once a pickup is completed. Admins schedule and complete pickups,
confirm payments with an uploaded proof file, and read aggregate reports.
Built with Go, PostgreSQL and MinIO. Payment proofs are stored in MinIO, an
S3-compatible object store that runs in docker as the local alternative to
AWS S3, so the whole stack works offline without any cloud account.

## Requirements

- Docker with the Compose plugin (`docker compose`)
- make
- curl, only for the simulation script

Go 1.24+ is only needed if you want to run the app or its tests outside
docker; the compose build compiles the binary inside the image.

## Quick start

```bash
cp .env.example .env
make up
make simulate
```

`make up` wraps `docker compose up --build -d` and starts the API, PostgreSQL
and MinIO. It also runs `migrate up` for you: a dedicated `migrate` service
applies every pending migration and must finish before the app container is
allowed to start, so there is no separate migrate step to remember.
`make simulate` then drives every flow through the real API to produce sample
data, see the simulation section below. Stop everything with `make down`.

The API listens on `http://localhost:8080`. Import
`community-waste.postman_collection.json` into Postman to click through every
endpoint; the simulation prints ready-to-paste values for its id variables
when it finishes.

The MinIO console is available at `http://localhost:9001` (log in with
`minioadmin` / `minioadmin` on local). Use it to browse the `payment-proofs`
bucket and preview the uploaded proof files.

## Environment variables

All variables are required. The app fails fast at startup and lists every
missing key, so a bad deploy dies immediately instead of limping along with
silent defaults. Copy `.env.example` to `.env` and adjust as needed.

| Variable | Example | Purpose |
|---|---|---|
| APP_PORT | 8080 | HTTP listen port |
| LOG_LEVEL | info | log level: debug, info, warn, error |
| DB_HOST | localhost | Postgres host (compose overrides to `postgres`) |
| DB_PORT | 5432 | Postgres port |
| DB_USER | postgres | Postgres user |
| DB_PASSWORD | postgres | Postgres password |
| DB_NAME | community_waste | database name |
| DB_SSLMODE | disable | Postgres sslmode |
| S3_ENDPOINT | localhost:9000 | MinIO endpoint (compose overrides to `minio:9000`) |
| S3_PUBLIC_ENDPOINT | localhost:9000 | host used when building stored proof URLs |
| S3_ACCESS_KEY | minioadmin | MinIO access key |
| S3_SECRET_KEY | minioadmin | MinIO secret key |
| S3_BUCKET | payment-proofs | bucket for payment proofs, created at startup if missing |
| S3_USE_SSL | false | use TLS for MinIO |
| SHUTDOWN_TIMEOUT | 30s | graceful shutdown budget |
| AUTOCANCEL_INTERVAL | 10s | organic auto-cancel worker tick |
| AUTOCANCEL_MAX_AGE | 72h | age after which unscheduled organic pickups are canceled |
| PICKUP_RATE_LIMIT_RPS | 1 | pickup creation rate limit refill per IP |
| PICKUP_RATE_LIMIT_BURST | 5 | pickup creation burst per IP |

The MinIO credentials above are compose-local defaults for development, not
real secrets. Nothing sensitive lives in the repository.

## Migrations and simulation

Migrations live in `migrations/` and are applied automatically by `make up`:
its compose stack includes a `migrate` service (golang-migrate) that runs
`migrate up` to completion before the app container starts. That is why there
is no manual migrate-up command anywhere in this README, booting the stack
covers it. The only migration command you might run by hand is the rollback:

```bash
make migrate-down
```

There is no SQL seed. Sample data comes from `scripts/simulate.sh`, which
walks every flow through the real endpoints, so whatever ends up in the
database is the genuine result of the API doing its job:

```bash
make simulate
```

The script hits every endpoint of the API at least once and asserts every
response status, success and failure alike, so a passing run is effectively
an integration test of the whole system; it exits nonzero on the first
unexpected answer.

The story it plays out: three households are registered, listed and fetched.
A fourth, Dewi, is registered and deleted again to show that a household
without records can be removed and is really gone afterwards (404). Budi gets
a paper pickup scheduled, completed and its auto-created payment confirmed
with a real proof upload to MinIO, plus a plastic pickup left scheduled whose
fee he pays up front through a manual payment entry, also confirmed, and an
electronic pickup waiting with its safety check. Sari completes a pickup
whose auto-created payment stays pending, and the script proves that this
blocks her next request (422). Andi keeps a fresh organic request and cancels
another. The organic auto-cancel rule is demonstrated too: the script creates
one more organic request through the API, backdates only that row's
`created_at` in the database (the passage of time is the one thing the API
cannot fake), and waits for the real worker to cancel it; this is why
`.env.example` ships with a 10 second `AUTOCANCEL_INTERVAL`, the 72 hour age
rule itself is untouched.

The failure paths covered along the way: an electronic request without
`safety_check` (400), scheduling a canceled pickup (409), confirming an
already paid payment (409), deleting a household that still has pickups and
payments (409), recording a payment against someone else's pickup (422),
asking for the history of an unknown household (404), and hammering pickup
creation until the rate limiter answers 429.

The script needs curl, plus the local compose stack for the auto-cancel
demonstration (that one step reaches into Postgres with `docker compose exec`
to backdate the timestamp). Rerunning it appends a fresh cast of the same
households, ids are generated per run.

## Architecture

Domain-first modules, each one Go package holding its whole vertical slice
(entity, handler, service, repo, raw SQL):

```
cmd/
  main.go          start, graceful shutdown orchestration
  inject.go        all dependency wiring (the only place domains meet)
internal/
  household/       handler.go service.go repo.go query.go entity.go
  pickup/          handler.go service.go repo.go query.go entity.go worker.go
  payment/         handler.go service.go repo.go query.go entity.go
  report/          handler.go service.go repo.go query.go entity.go
  server/          router.go middleware.go (rate limit, request logging)
external/
  storage/         FileStorage interface and MinIO implementation
pkg/
  config/          env loading, fail fast on missing keys
  db/              sqlx wrapper, tx-in-context helper
  apperr/          AppError {Code, Message}
  httpres/         response envelope helpers, error mapping, pagination
  logger/          leveled logger
scripts/
  simulate.sh      drives every API flow to produce sample data
```

Domain imports follow a declared tree and only point down it:

```
household   leaf, imports no domain
   ^
pickup      imports household
   ^
payment     imports pickup and household

report      imports no domain, owns its aggregate SQL
```

Exactly one relationship goes against the tree: pickup needs payment for the
pending-payment check and for creating the payment on completion. Pickup
declares that interface itself (`pickup.PaymentService`) using its own types,
the payment service satisfies it implicitly, and `cmd/inject.go` wires them
together. Upward needs always cross through an interface declared by the
consumer, so no import cycle is possible.

All business rules live in the service layer. Handlers only parse and respond,
repositories only run SQL. Completing a pickup must update the pickup and
create its payment atomically, so `pkg/db` provides `WithTx`, which stores the
transaction in the context; repositories resolve their executor from the
context (transaction if present, pool otherwise), keeping `*sqlx.Tx` out of
every interface.

The report module is a read model: it owns its aggregate SQL directly and
never reaches into the other domains' repositories, and writes always go
through the owning domain.

## API overview

Base path `/api`. All ids are UUIDs.

| Method | Path | Description | Errors |
|---|---|---|---|
| GET | /health | liveness check | |
| POST | /api/households | register household | 400 |
| GET | /api/households | list households, `page`, `limit` | |
| GET | /api/households/{id} | household detail | 404 |
| DELETE | /api/households/{id} | delete household | 404, 409 if it still has pickups or payments |
| POST | /api/pickups | request pickup | 400, 404, 422 pending payment or missing safety_check, 429 |
| GET | /api/pickups | list, filters `status`, `household_id`, `page`, `limit` | 400 |
| PUT | /api/pickups/{id}/schedule | set pickup date | 400, 404, 409 not pending, 422 electronic without safety check |
| PUT | /api/pickups/{id}/complete | complete and auto-create payment | 404, 409 not scheduled |
| PUT | /api/pickups/{id}/cancel | cancel pickup | 404, 409 already completed or canceled |
| POST | /api/payments | record payment | 400, 404, 422 |
| GET | /api/payments | list, filters `status`, `household_id`, `date_from`, `date_to`, `page`, `limit` | 400 |
| PUT | /api/payments/{id}/confirm | confirm with proof upload (multipart, file key `proof`) | 400, 404, 409 not pending |
| GET | /api/reports/waste-summary | pickup counts grouped by type and status | |
| GET | /api/reports/payment-summary | payment counts and revenue by status | |
| GET | /api/reports/households/{id}/history | full pickup and payment history | 404 |

Every response uses one envelope. Success:

```json
{
  "code": 200,
  "message": "success",
  "data": { "id": "a1111111-1111-1111-1111-111111111111", "owner_name": "Budi Santoso" }
}
```

List responses add a `meta` object with `page`, `limit` and `total`. Errors
carry no data:

```json
{
  "code": 422,
  "message": "household has a pending payment"
}
```

## Business rules

1. A household with a pending payment cannot request a new pickup (422).
2. Only pending pickups can be scheduled; anything else is a state conflict (409).
3. Electronic pickups must include `safety_check` at creation (400 if absent)
   and can only be scheduled when it is true (422 otherwise).
4. Organic pickups left pending or scheduled longer than `AUTOCANCEL_MAX_AGE`
   are canceled by a background worker in `internal/pickup/worker.go`; it runs
   on a ticker and exits cleanly when the shutdown context is canceled.
5. Completing a pickup creates its payment in the same transaction: pending
   status, 50,000 for organic, plastic and paper, 100,000 for electronic.
6. Confirming a payment uploads the proof file to MinIO first, then marks the
   payment paid and sets the payment date.

Rules 1 and 5 chain together: completing a pickup creates a pending payment
that blocks that household's next request until it is confirmed. The
simulation demonstrates exactly this with Sari's household.

## Testing

```bash
make test
```

The service layer is unit tested with generated mocks of the repositories,
neighbor services and file storage. Mocks are generated by mockgen into
`test/mocks/` and regenerated with `make mock` (`make test` runs it first).

## Design decisions and tradeoffs

- Foreign keys are `ON DELETE RESTRICT`, so deleting a household with history
  returns 409 instead of cascading or soft-deleting. History stays intact and
  the client gets an honest answer.
- Enums (pickup type and status, payment status) are text columns with CHECK
  constraints rather than native Postgres enum types. Native enums are painful
  to evolve: `ALTER TYPE ... ADD VALUE` has transaction restrictions, and
  removing or renaming a value is not supported at all, it takes creating a
  replacement type, casting every column over and dropping the old one. With
  a CHECK constraint, any change is an ordinary transactional migration that
  drops the constraint and recreates it with the new list, and the down
  migration just restores the previous list. The constraint also sits right
  next to the column in the schema, so the allowed values are readable in one
  place. The cost is a few bytes per row and validation by string comparison
  instead of a compact enum id, which is irrelevant at this scale.
- Money is `decimal.Decimal` end to end, never float.
- Rate limiting is a per-IP token bucket applied only to pickup creation, the
  single endpoint residents can hammer. Exceeding it returns 429. Idle IP
  entries are evicted by a janitor goroutine.
- On payment confirmation the file upload happens before the DB update. If the
  DB update then fails, the orphaned object in MinIO is logged and accepted;
  an orphan file is harmless while a paid record without a proof is not.
