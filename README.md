# Parlor API Core

[![Go Test](https://github.com/parlorhub/api-core/actions/workflows/go-test.yml/badge.svg)](https://github.com/parlorhub/api-core/actions/workflows/go-test.yml)
[![Semantic Release](https://github.com/parlorhub/api-core/actions/workflows/semantic-release.yml/badge.svg)](https://github.com/parlorhub/api-core/actions/workflows/semantic-release.yml)
[![Release](https://github.com/parlorhub/api-core/actions/workflows/release.yml/badge.svg)](https://github.com/parlorhub/api-core/actions/workflows/release.yml)
[![Commit Lint](https://github.com/parlorhub/api-core/actions/workflows/commit-lint.yml/badge.svg)](https://github.com/parlorhub/api-core/actions/workflows/commit-lint.yml)
![Go Version](https://img.shields.io/github/go-mod/go-version/parlorhub/api-core)
[![GitHub release](https://img.shields.io/github/v/release/parlorhub/api-core)](https://github.com/parlorhub/api-core/releases)
[![License](https://img.shields.io/github/license/parlorhub/api-core)](LICENSE)

Backend API for the Parlor salon management platform. Handles authentication, appointment booking, client management, onboarding, and more.

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.25 |
| Framework | Gin |
| Database | PostgreSQL 18 |
| Cache | Redis 7.4 |
| Auth | JWT + Google OAuth2 (SSO) |
| Email | Mailtrap |
| SQL Generation | sqlc |
| Docs | OpenAPI 3.1 / Swagger UI |
| Containers | Docker & Docker Compose |
| Testing | testcontainers (real PostgreSQL) |
| Release | GoReleaser + Semantic Release |

## Project Structure

```
cmd/api/main.go                  # Entry point
internal/
├── server/                      # Server init & route registration
├── modules/
│   ├── auth/                    # Registration, login, JWT tokens
│   ├── sso/                     # Google OAuth2 sign-in
│   ├── profile/                 # User profile CRUD
│   ├── onboarding/              # Onboarding flow & step tracking
│   └── contact_form/            # Contact form submissions
├── middleware/
│   ├── auth/                    # JWT authentication guard
│   ├── rbac/                    # Role-based access control
│   └── origin/                  # Origin validation
├── database/
│   └── sqlc/                    # Generated type-safe query code
├── logger/                      # Structured logging
├── email/                       # Mailtrap email service
└── helper/                      # Shared utilities
migrations/
├── schema/                      # Database schema migrations
└── queries/                     # SQL queries (sqlc input)
docs/                            # OpenAPI spec & Swagger UI
config/nginx/                    # File storage nginx config
test/                            # Integration & unit tests
```

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/health` | Health check | No |
| `POST` | `/auth/register` | Register user | No |
| `POST` | `/auth/login` | Login | No |
| `POST` | `/auth/refresh` | Refresh JWT token | No |
| `GET` | `/auth/me` | Current user info | Yes |
| `POST` | `/auth/logout` | Logout | Yes |
| `GET` | `/auth/google` | Start Google OAuth | No |
| `GET` | `/auth/google/callback` | Google OAuth callback | No |
| `GET` | `/profile` | Get profile | Yes |
| `PATCH` | `/profile` | Update profile | Yes |
| `GET` | `/onboarding` | Get onboarding progress | Yes |
| `POST` | `/onboarding` | Save onboarding selections | Yes |
| `POST` | `/onboarding/step` | Save individual step | Yes |
| `POST` | `/contact-form` | Submit contact form | Origin |
| `GET` | `/contact-form` | List submissions | Yes |
| `GET` | `/swagger` | Swagger UI | No |

## Database Schema

Core tables: `users`, `salons`, `sso`, `services`, `clients`, `appointments`, `notifications`, `contact_form`, `onboarding_steps`, `onboarding_options`, `salon_onboarding_selection`.

User roles: `admin`, `owner`, `employee`, `customer`
Appointment statuses: `pending`, `confirmed`, `cancelled`, `done`, `no-show`

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) (for manual migrations)

### Setup

```bash
# Clone and configure
cp .env.example .env
# Edit .env with your values

# Run with Docker (PostgreSQL + Redis + app + nginx)
make docker-run

# Or run locally
make run

# Hot reload (requires air)
make watch
```

### Database Migrations

```bash
make migrate-up          # Apply all migrations
make migrate-down        # Rollback last migration
make migrate-down-all    # Rollback all migrations
make migrate-create name=my_migration  # Create new migration
make migrate-status      # Check current version
```

### Testing

```bash
make test    # Runs integration tests with testcontainers
```

### Code Generation

```bash
make queries_gen    # Regenerate sqlc code from SQL queries
```

## Environment Variables

See [`.env.example`](.env.example) for all required variables. Key groups:

- **Database**: `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- **JWT**: `JWT_SECRET` (min 32 chars in production), `JWT_EXPIRATION_HOURS`
- **Google OAuth**: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- **Email**: `MAILTRAP_TOKEN`, `MAILTRAP_URL`
- **URLs**: `FRONTEND_URL`, `LANDING_PAGE_URL`

Use `.env.local` to override values for local development without modifying `.env`.

---

## Branches

### Active Feature Branches

#### `feat/dashboard`
Adds the salon dashboard and updates the onboarding endpoint. Builds on top of the management module work. Includes Swagger documentation updates.

**Key commits:**
- `111ebf5` feat: add dashboard and updated onboarding endpoint
- `86bc42f` fix: remove unnecessary doc

#### `feat/managment`
Introduces the salon/service management module — CRUD operations for salons, services, and clients. Adds Swagger API documentation integration.

**Key commits:**
- `9fc553d` feat: add managment
- `29bdec7` feat: add in swagger
- `f37a3e5` fix: fix managment packages structure and swagger

#### `development`
Integration branch. Currently has the management module merged ahead of main.

#### `feat/swagger`
Adds Swagger UI documentation endpoint to the API.

**Key commit:**
- `4a8bc25` feat: added swagger

### Merged / Historical Branches

These branches have been merged into `main` and represent completed work:

#### `fix/merge-sql`
Fixed SQL migration conflicts after multiple feature branches were merged. Consolidated migration files.

#### `ci/docker-artifacts`
Added CI workflow for building and publishing Docker image artifacts via GitHub Actions.

#### `feat/onboarding`
Added the user onboarding flow — multi-step wizard with selectable options stored per salon.

#### `feat/sso`
Implemented Google OAuth2 Single Sign-On. Added SSO database table, OAuth flow handlers, and token exchange.

#### `feat/rbac`
Added Role-Based Access Control middleware. Restricts endpoints based on user roles (`admin`, `owner`, `employee`, `customer`).

#### `feat/contact-form`
Added the contact form submission endpoint with origin-based access restriction (landing page only).

#### `feat/email-handler`
Integrated Mailtrap email service for transactional emails (registration confirmation, notifications).

#### `feat/move-to-uuidv7`
Migrated primary keys from default UUIDs to UUIDv7 for time-sortable, index-friendly identifiers.

#### `feat!/ci`
Breaking change — overhauled CI/CD pipeline with semantic release, GoReleaser, and conventional commit linting.

---

## Releases

| Version | Highlights |
|---------|-----------|
| **v2.2.1** | Merged SQL migration fixes, CI ref fixes |
| **v2.2.0** | Onboarding module, Docker artifact CI, SSO + RBAC merge into main |
| **v2.1.0** | Nginx port mapping via environment variable |
| **v2.0.0** | SSO (Google OAuth), RBAC, logger, contact form handler |
| **v1.x** | Initial auth system, base API structure, early features |

---

## Contributing

This project uses [Conventional Commits](https://www.conventionalcommits.org/). Commit messages are linted via CI.

```
feat: add new endpoint        → minor version bump
fix: correct token validation → patch version bump
feat!: redesign auth flow     → major version bump
```
