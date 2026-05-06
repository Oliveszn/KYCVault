# KYCVault

A production-grade Know Your Customer (KYC) verification platform built with **Go** (Gin) and **React**. Users register, submit identity documents, upload a selfie, and an admin manually reviews and approves or rejects the verification. Real-time status updates are pushed to connected clients over WebSocket.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Database Schema](#database-schema)
- [KYC Session State Machine](#kyc-session-state-machine)
- [API Reference](#api-reference)
- [Authentication](#authentication)
- [Real-time Notifications](#real-time-notifications)
- [Audit Log](#audit-log)
- [Security Decisions](#security-decisions)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)

---

## Overview

KYCVault walks a user through a five-step verification wizard:

1. **Register / Login** — email, name, password
2. **Initiate session** — select country and ID type
3. **Document upload** — front and back of the identity document (JPEG, PNG, or PDF, max 10 MB each)
4. **Selfie upload** — a single photo taken via the device camera
5. **Admin review** — a human reviewer approves or rejects; the user receives a real-time notification either way

Supported ID types: `national_id`, `drivers_license`, `passport`, `residence_permit`.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│                React Frontend               │
│  Registration · Wizard · Status · WebSocket │
└───────────────────┬─────────────────────────┘
                    │ HTTPS / WSS
┌───────────────────▼─────────────────────────┐
│              Go API (Gin)                   │
│  JWT auth · Rate limiting · Routing         │
├──────────┬──────────┬──────────┬────────────┤
│   Auth   │   KYC    │ Document │    Face    │
│ Service  │ Service  │ Service  │  Service   │
├──────────┴──────────┴──────────┴────────────┤
│          Notification Service               │
│       WebSocket Hub · DB notifications      │
├─────────────────────────────────────────────┤
│              Audit Service                  │
│          Append-only event log              │
├──────────┬──────────┬──────────┬────────────┤
│  users   │  kyc_    │  kyc_    │   face_    │
│  refresh │ sessions │documents │verification│
│  tokens  │          │          │            │
├──────────┴──────────┴──────────┴────────────┤
│             PostgreSQL · S3/GCS             │
└─────────────────────────────────────────────┘
```

The backend follows a strict three-layer pattern throughout:

```
Handler  →  Service  →  Repository  →  Database
```

Handlers parse HTTP input and map errors to status codes. Services own all business logic. Repositories own all GORM queries. Nothing crosses layer boundaries — a handler never touches GORM, a repository never contains business logic.

---

## Tech Stack

| Layer            | Technology                                   |
| ---------------- | -------------------------------------------- |
| Backend language | Go 1.22+                                     |
| HTTP framework   | Gin                                          |
| ORM              | GORM                                         |
| Database         | PostgreSQL 15+                               |
| Object storage   | S3 / GCS (via `storage.Client` interface)    |
| Real-time        | WebSocket (gorilla/websocket)                |
| Logging          | Uber Zap                                     |
| Auth tokens      | JWT (access) + SHA-256 hashed refresh tokens |
| Password hashing | bcrypt                                       |
| Frontend         | React                                        |

---

## Project Structure

```
kycvault/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point — wires all dependencies
├── internal/
│   ├── dtos/
│   │   └── kyc_dto.go               # Request/response shapes
│   ├── handlers/
│   │   ├── helpers.go               # respond(), respondError(), parsePagination()
│   │   ├── kyc_handler.go           # Session lifecycle + admin review endpoints
│   │   ├── document_handler.go      # Multipart upload + document listing
│   │   ├── face_handler.go          # Selfie upload
│   │   └── webhook_handler.go       # Webhook endpoint management
│   ├── services/
│   │   ├── audit_service.go         # Append-only event logger
│   │   ├── kyc_service.go           # State machine — the core of the platform
│   │   ├── document_service.go      # File validation, storage, session advancement
│   │   ├── face_service.go          # Selfie storage, verification submission
│   │   ├── notification_service.go  # WebSocket push + DB notification persistence
│   │   └── webhook_service.go       # Endpoint registration, delivery, retry
│   ├── repository/
│   │   ├── audit_repository.go
│   │   ├── kyc_repository.go
│   │   ├── document_repository.go
│   │   ├── face_repository.go
│   │   └── webhook_repository.go
│   ├── models/
│   │   ├── user.go
│   │   ├── refresh_token.go
│   │   ├── kyc_session.go
│   │   ├── kyc_document.go
│   │   ├── face_verification.go
│   │   ├── notification.go
│   │   └── audit_event.go
│   ├── storage/
│   │   └── client.go                # Storage interface + key builders
│   ├── middleware/
│   │   └── auth.go                  # JWT extraction, MustGetUserID()
│   └── workers/
│       └── webhook_worker.go        # Background delivery retry loop
└── migrations/
    └── schema.go                    # Indexes, constraints, append-only trigger
```

---

## Database Schema

### Core tables

**`users`**
Stores account credentials and profile. `password_hash` is bcrypt. `role` is `user` or `admin`.

**`refresh_tokens`**
Single-use, rotation-based. Stores only the SHA-256 hash of the raw token — the raw value is never written to the database. Revoked tokens are retained for audit and reuse-detection (a revoked token being presented signals theft and triggers full session termination).

**`kyc_sessions`**
The root record for one verification attempt. Owns the state machine. A partial unique index (`WHERE status NOT IN ('approved', 'rejected')`) ensures a user can only have one active session at a time at the database level — not just application level.

| Column             | Type        | Notes                                                                  |
| ------------------ | ----------- | ---------------------------------------------------------------------- |
| `id`               | UUID        | PK                                                                     |
| `user_id`          | UUID        | FK → users                                                             |
| `status`           | text        | See state machine below                                                |
| `country`          | char(2)     | ISO 3166-1 alpha-2                                                     |
| `id_type`          | text        | `national_id` \| `drivers_license` \| `passport` \| `residence_permit` |
| `reviewer_id`      | UUID        | FK → users (admin)                                                     |
| `review_note`      | text        | Internal only                                                          |
| `rejection_reason` | text        | User-facing                                                            |
| `attempt_number`   | int         | Increments on re-submission                                            |
| `expires_at`       | timestamptz | 24h TTL from creation                                                  |

**`kyc_documents`**
One row per uploaded image (front or back). Raw bytes are never stored — only the S3/GCS object key, bucket name, MIME type, file size, and a SHA-256 checksum for integrity verification. A partial unique index ensures at most one accepted document per side per session.

**`face_verifications`**
One row per session (upserted on retry). Stores liveness score, match score, pass/fail flags, the selfie storage key, and the raw vendor response. Since verification is manual in the current implementation, the row records the selfie location and awaits admin decision.

**`notifications`**
Persisted notifications for users. Each row has a `read` flag. The WebSocket hub pushes the same payload in real time; the DB copy allows users to retrieve missed notifications on reconnect.

**`audit_events`**
Append-only. A Postgres trigger (`trg_audit_events_no_update`, `trg_audit_events_no_delete`) prevents any UPDATE or DELETE at the database level — independent of application code. Every status transition, document upload, admin action, and auth event produces a row here.

### Key indexes

```sql
-- Prevents two active sessions per user (enforced at DB, not just app level)
CREATE UNIQUE INDEX idx_kyc_sessions_one_active_per_user
  ON kyc_sessions (user_id)
  WHERE status NOT IN ('approved', 'rejected');

-- Prevents accepting two images for the same side
CREATE UNIQUE INDEX idx_kyc_documents_one_accepted_per_side
  ON kyc_documents (session_id, side)
  WHERE status = 'accepted';

-- Retry worker polling query
CREATE INDEX idx_webhook_deliveries_retry_queue
  ON webhook_deliveries (next_retry_at)
  WHERE status = 'pending' AND next_retry_at IS NOT NULL;
```

---

## KYC Session State Machine

Every status transition is validated by `KYCService.AdvanceStatus` against a compile-time map. Nothing in the codebase writes to the `status` column except through this function.

```
INITIATED ──► DOC_UPLOAD ──► FACE_VERIFY ──► IN_REVIEW ──► APPROVED
                                                    │
                                                    └──────► REJECTED
```

| Transition                 | Triggered by                           |
| -------------------------- | -------------------------------------- |
| `initiated → doc_upload`   | First document uploaded                |
| `doc_upload → face_verify` | Both front and back documents accepted |
| `face_verify → in_review`  | Selfie uploaded                        |
| `in_review → approved`     | Admin manually approves                |
| `in_review → rejected`     | Admin manually rejects                 |
| Any → `rejected`           | Session TTL expires (24h)              |

When a session reaches `approved` or `rejected`, the notification service pushes a real-time WebSocket message to all of the user's connected tabs and persists a notification row to the database.

---

## API Reference

All endpoints are prefixed with `/api/v1`. Authenticated routes require `Authorization: Bearer <access_token>`.

### Auth

| Method | Path               | Auth   | Description                        |
| ------ | ------------------ | ------ | ---------------------------------- |
| POST   | `/auth/register`   | —      | Create account                     |
| POST   | `/auth/login`      | —      | Login, sets `refresh_token` cookie |
| POST   | `/auth/refresh`    | Cookie | Rotate access + refresh tokens     |
| POST   | `/auth/logout`     | Cookie | Revoke current refresh token       |
| POST   | `/auth/logout-all` | JWT    | Revoke all sessions for the user   |

**Register body**

```json
{
  "email": "user@example.com",
  "first_name": "Ada",
  "last_name": "Okonkwo",
  "password": "...",
  "confirm_password": "..."
}
```

**Login response**

```json
{
  "message": "login successful",
  "data": {
    "accessToken": "eyJ...",
    "expiresIn": 900,
    "tokenType": "Bearer"
  }
}
```

The refresh token is set as an `httpOnly`, `Secure`, `SameSite=Strict` cookie — never in the response body.

---

### KYC Sessions

| Method | Path                    | Auth | Description                                |
| ------ | ----------------------- | ---- | ------------------------------------------ |
| POST   | `/kyc/sessions`         | JWT  | Initiate a new KYC session                 |
| GET    | `/kyc/sessions/active`  | JWT  | Get current in-progress session            |
| GET    | `/kyc/sessions/history` | JWT  | All sessions for the user                  |
| GET    | `/kyc/sessions/:id`     | JWT  | Get a specific session (ownership-checked) |

**Initiate session body**

```json
{
  "country": "NG",
  "id_type": "national_id"
}
```

**Session response**

```json
{
  "id": "550e8400-...",
  "status": "initiated",
  "country": "NG",
  "id_type": "national_id",
  "attempt_number": 1,
  "expires_at": "2024-01-02T15:04:05Z",
  "documents": []
}
```

---

### Documents

| Method | Path                           | Auth        | Description                   |
| ------ | ------------------------------ | ----------- | ----------------------------- |
| POST   | `/kyc/sessions/:id/documents`  | JWT         | Upload front or back document |
| GET    | `/kyc/sessions/:id/documents`  | JWT         | List documents for a session  |
| GET    | `/admin/documents/:doc_id/url` | JWT + Admin | Get 15-min presigned view URL |

**Upload** — `multipart/form-data`

| Field  | Type   | Required | Notes                         |
| ------ | ------ | -------- | ----------------------------- |
| `side` | string | ✓        | `front` or `back`             |
| `file` | file   | ✓        | JPEG, PNG, or PDF · max 10 MB |

Uploading the first document advances the session from `initiated` → `doc_upload`. Uploading both sides advances it to `face_verify` automatically.

---

### Face Verification

| Method | Path                     | Auth | Description             |
| ------ | ------------------------ | ---- | ----------------------- |
| POST   | `/kyc/sessions/:id/face` | JWT  | Upload selfie           |
| GET    | `/kyc/sessions/:id/face` | JWT  | Get verification status |

**Upload** — `multipart/form-data`

| Field  | Type | Required | Notes           |
| ------ | ---- | -------- | --------------- |
| `file` | file | ✓        | JPEG · max 5 MB |

Uploading the selfie stores the image and advances the session to `in_review`, queuing it for admin review.

---

### Admin

| Method | Path                              | Auth        | Description                        |
| ------ | --------------------------------- | ----------- | ---------------------------------- |
| GET    | `/admin/kyc/sessions`             | JWT + Admin | Paginated review queue             |
| GET    | `/admin/kyc/sessions/counts`      | JWT + Admin | Count per status (dashboard tiles) |
| GET    | `/admin/kyc/sessions/:id`         | JWT + Admin | Full session detail                |
| POST   | `/admin/kyc/sessions/:id/approve` | JWT + Admin | Approve a session                  |
| POST   | `/admin/kyc/sessions/:id/reject`  | JWT + Admin | Reject a session                   |

**Approve body**

```json
{ "note": "Documents verified. Face matches." }
```

**Reject body**

```json
{
  "note": "Internal: document edge cut off.",
  "reason": "The document image was incomplete. Please re-upload ensuring all four corners are visible."
}
```

`note` is internal-only (stored in audit log). `reason` is user-facing (stored on the session and sent in the real-time notification).

---

### Notifications

| Method | Path                      | Auth | Description                          |
| ------ | ------------------------- | ---- | ------------------------------------ |
| GET    | `/notifications`          | JWT  | Fetch all notifications for the user |
| PATCH  | `/notifications/:id/read` | JWT  | Mark a notification as read          |

---

### WebSocket

```
GET /ws?token=<access_token>
```

Upgrades to a WebSocket connection. The server pushes a JSON message to all connected tabs for the user whenever a KYC event occurs.

**Notification payload**

```json
{
  "type": "notification",
  "data": {
    "id": "...",
    "user_id": "...",
    "title": "Identity Verified",
    "body": "Your KYC verification has been approved.",
    "read": false,
    "created_at": "2024-01-02T15:04:05Z"
  }
}
```

The Hub handles multiple tabs per user — each connected `*websocket.Conn` for a given `user_id` receives the push. Dead connections are unregistered automatically on write error.

---

## Authentication

**Access tokens** — short-lived JWT (15 minutes). Signed with HMAC-SHA256. Claims carry `user_id` and `role`.

**Refresh tokens** — long-lived (7 days), stored in an `httpOnly` cookie. The raw token is a 32-byte cryptographically random hex string. Only the SHA-256 hash is ever written to the database. On every `/auth/refresh` call the old token is immediately revoked and a new one is issued (rotation).

**Reuse detection** — presenting a previously revoked refresh token is treated as a signal of token theft. All refresh tokens for that user are immediately revoked, forcing a full re-login on all devices.

---

## Real-time Notifications

```
Admin approves session
        │
        ▼
KYCService.ApproveSession()
        │
        ├──► NotificationService.Create()
        │         │
        │         ├──► repo.Create()     — persists to DB
        │         └──► hub.SendToUser()  — pushes to all user WS connections
        │
        └──► [fire-and-forget goroutine]
```

The notification service never blocks the admin's HTTP response. Both the DB write and the WebSocket push happen in a goroutine. If the WebSocket write fails (client disconnected), the connection is unregistered and the notification remains readable via `GET /notifications` on reconnect.

---

## Audit Log

Every meaningful event in the system writes a row to `audit_events`. The table is append-only, enforced by Postgres triggers that raise an exception on any UPDATE or DELETE.

| Event type                                  | Triggered by                   |
| ------------------------------------------- | ------------------------------ |
| `auth.login_success` / `auth.login_failure` | Auth handler                   |
| `auth.token_revoked`                        | Refresh / logout               |
| `session.created`                           | KYCService.InitiateSession     |
| `session.status_changed`                    | KYCService.AdvanceStatus       |
| `document.uploaded`                         | DocumentService.UploadDocument |
| `document.accepted` / `document.rejected`   | Document validation            |
| `face_verify.started`                       | FaceService.StartVerification  |
| `admin.manual_approval`                     | KYCService.ApproveSession      |
| `admin.manual_rejection`                    | KYCService.RejectSession       |

Each row carries: actor ID + role, session ID, user ID, event type, free-form JSONB metadata, IP address, user agent, and request correlation ID.

---

## Security Decisions

**Passwords** — bcrypt at `DefaultCost`. Timing-safe comparison via `bcrypt.CompareHashAndPassword`.

**Refresh tokens** — never stored in plaintext. SHA-256 hashed at rest. Rotation on every use. Full revocation on reuse detection.

**Document storage** — raw files never stored in Postgres. Only the S3/GCS object key, bucket, MIME type, size, and SHA-256 checksum are stored. Images are accessible only via short-lived presigned URLs (15 minutes), admin-only.

**Ownership checks** — every user-facing endpoint that accepts a session ID calls `GetSessionForUser`, which returns `404 Not Found` (not `403 Forbidden`) if the session belongs to a different user. This prevents enumeration — users cannot discover that a session ID exists.

**Partial unique indexes** — business-critical constraints (one active session per user, one accepted document per side) are enforced at the database level, not just application level. A bug in the service layer cannot create corrupt state.

**Append-only audit log** — Postgres triggers prevent any UPDATE or DELETE on `audit_events` even if application code is compromised.

**Input validation** — file type is validated from both the `Content-Type` header and the file extension. File size is capped before bytes are read into memory.

---

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- An S3-compatible object store (AWS S3, GCS with interop, or MinIO for local dev)

### Run locally

```bash
# 1. Clone
git clone https://github.com/yourorg/kycvault.git
cd kycvault

# 2. Copy and fill in environment variables
cp .env.example .env

# 3. Create the database
createdb kycvault

# 4. Run migrations (AutoMigrate + raw SQL schema runs on startup)
go run ./cmd/server

# 5. The server starts on :8080
```

---

## Environment Variables

| Variable                | Example                              | Description                     |
| ----------------------- | ------------------------------------ | ------------------------------- |
| `DATABASE_URL`          | `postgres://localhost:5432/kycvault` | Postgres connection string      |
| `JWT_SECRET`            | `changeme-32-bytes-min`              | HMAC signing key for JWTs       |
| `JWT_ACCESS_TTL`        | `15m`                                | Access token lifetime           |
| `JWT_REFRESH_TTL`       | `168h`                               | Refresh token lifetime (7 days) |
| `STORAGE_BUCKET`        | `kycvault-documents`                 | S3/GCS bucket name              |
| `STORAGE_REGION`        | `us-east-1`                          | AWS region                      |
| `AWS_ACCESS_KEY_ID`     | —                                    | S3 credentials                  |
| `AWS_SECRET_ACCESS_KEY` | —                                    | S3 credentials                  |
| `COOKIE_DOMAIN`         | `kycvault.io`                        | Refresh token cookie domain     |
| `COOKIE_SECURE`         | `true`                               | Set `false` for local HTTP dev  |
| `PORT`                  | `8080`                               | HTTP listen port                |
