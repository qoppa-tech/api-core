# API Usage Instructions

## Overview
- Base URL (local): `http://localhost:8080`
- Authentication: JWT (access + refresh) via cookies or `Authorization: Bearer <token>` header.
- Protected routes require `AuthMiddleware`; some also require RBAC (owner/admin).
- Swagger UI: `GET /docs` → redirects to `/swagger/swagger.html`. OpenAPI spec at `docs/openapi.yaml`.

## Auth Flow
1) Register: `POST /auth/register` (email, password, name, phone, optional role).
2) Login: `POST /auth/login` → sets access/refresh cookies and returns user payload.
3) Me: `GET /auth/me` (auth required).
4) Refresh: `POST /auth/refresh` (uses refresh cookie).
5) Logout: `POST /auth/logout` (auth required) clears cookies.
6) SSO Google: `GET /auth/google`, callback at `/auth/google/callback`.

## Roles & RBAC
- Roles (ascending): customer < employee < owner < admin.
- Owner/admin required for salons and services mutations (RBAC middleware + ownership checks).
- Ownership: for salons/services, the authenticated user must own the salon (admins bypass).
- Clients and appointments require auth; role not enforced unless noted below.

## Entities & Endpoints

### Salons (owner/admin only for mutations)
- `POST /salons` create (owner_id auto from token).
- `GET /salons` list all.
- `GET /salons/by-slug?slug=` get by slug.
- `GET /salons/by-owner?owner_id=` get by owner.
- `GET /salons/:id` get by id.
- `PATCH /salons/:id` update (owner/admin, must own).
- `DELETE /salons/:id` delete (owner/admin, must own).

### Services (owner/admin only for mutations; must own salon)
- `POST /services` create (fields: salon_id, optional user_id, name, duration, price, active).
- `GET /services?salon_id=` list by salon.
- `GET /services/active?salon_id=` list active.
- `GET /services/by-user?user_id=` list by user.
- `GET /services/by-salon-and-user?salon_id=&user_id=` list assigned + unassigned for user.
- `GET /services/:id` get by id.
- `PATCH /services/:id` update (owner/admin, owns salon).
- `DELETE /services/:id` delete (owner/admin, owns salon).

### Clients (auth required; intended for external/AI integration)
- `POST /clients` create (salon_id, name, phone, optional email, birthday).
- `GET /clients?salon_id=` list by salon.
- `GET /clients/search?salon_id=&q=` search by name/phone.
- `GET /clients/by-phone?salon_id=&phone=` lookup by phone.
- `GET /clients/:id` get by id.
- `PATCH /clients/:id` update.
- `DELETE /clients/:id` delete.

### Appointments (auth required)
- `POST /appointments` create (salon_id, user_id, service_id, client_name/phone, optional email, date YYYY-MM-DD, start_time HH:MM, end_time HH:MM, optional status, notes).
- `GET /appointments?salon_id=&limit=&offset=` list by salon.
- `GET /appointments/by-date?salon_id=&date=` list by date.
- `GET /appointments/by-date-range?salon_id=&start_date=&end_date=` list by range.
- `GET /appointments/by-user?user_id=` list by user.
- `GET /appointments/by-status?salon_id=&status=` list by status.
- `GET /appointments/by-client-phone?salon_id=&phone=` list by client phone.
- `GET /appointments/count?user_id=&date=` count by user and date.
- `GET /appointments/:id` get by id.
- `PATCH /appointments/:id` update (date/times/status/notes).
- `PATCH /appointments/:id/status` update status only.
- `DELETE /appointments/:id` delete.

### Contact Form (public POST, protected list)
- `POST /contact-form` (public, origin-gated) submit form.
- `GET /contact-form` (auth) list submissions.

### Profile
- `GET /profile` (auth) get profile.
- `PATCH /profile` (auth) update profile.

### Onboarding
- `GET /onboarding` (auth) get steps + user progress.
- `POST /onboarding` (auth) save batch selections.
- `POST /onboarding/step` (auth) save single step (mark last step via `is_last_step`).

## Security Notes
- Auth middleware sets `user_id`, `email`, `role`, and optional `salon_id` on context.
- RBAC middleware compares required role vs user role.
- Salons/services handlers also enforce ownership server-side.

## Running Locally
- Start API (from repo root): `go run ./cmd/api` or `make run` (if present).
- Docker compose (db/redis): `docker compose up -d` (services defined in `docker-compose.yml`).

## Docs
- Swagger UI: `/docs` (served from `./docs`).
- OpenAPI file: `docs/openapi.yaml` (kept in sync with code).
