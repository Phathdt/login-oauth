# login-oauth

Full-stack authentication demo using Firebase Authentication with Go (Fiber) backend and React (TypeScript) frontend. Supports Google, GitHub, and Email/Password sign-in.

## Stack

| Layer | Tech |
|-------|------|
| Frontend | React 19 + Vite + TypeScript, React Router v7, TanStack Query v5, Axios, Tailwind CSS v4, shadcn/ui |
| Backend | Go + Fiber v2, Firebase Admin SDK, golang-jwt/jwt v5 |
| Database | PostgreSQL (via pgx/v5), sqlc, Goose migrations |
| Dev infra | Docker Compose (postgres only) |

## Project Structure

```
login-oauth/
├── web/                    # React frontend (port 5173)
│   └── src/
│       ├── contexts/       # Auth context (in-memory token state)
│       ├── lib/            # firebase.ts, axios-client.ts, query-client.ts
│       ├── routes/         # login, products
│       └── components/     # ProtectedRoute, Navbar, ProductCard
├── api/                    # Go backend (port 3000)
│   ├── cmd/server/         # Entry point
│   └── internal/
│       ├── auth/           # firebase.go, jwt.go, middleware.go
│       ├── handlers/       # auth_handler.go, product_handler.go
│       ├── models/         # refresh_token.go
│       ├── config/         # env config loader
│       └── database/       # pgx pool + Goose migrations
├── docker-compose.yml
└── README.md
```

## Login Flow

```
Browser              React (5173)              Firebase             Go Backend (3000)
  │                      │                        │                       │
  │── visit /login ─────>│                        │                       │
  │── click provider ───>│                        │                       │
  │                      │── signInWithPopup() ──>│                       │
  │                      │<── FirebaseUser ────── │                       │
  │                      │   + id_token (1h)      │                       │
  │                      │                        │                       │
  │                      │── POST /auth/firebase ─────────────────────── >│
  │                      │   { id_token }         │    VerifyIDToken() ──>│
  │                      │                        │<── decoded claims ────│
  │                      │                        │       upsert user ────│
  │                      │                        │    generate JWT (15m) │
  │                      │                        │  generate refresh (7d)│
  │                      │<── { access_token, user } + HTTP-only cookie ──│
  │                      │                        │                       │
  │                      │ store token in memory  │                       │
  │<── /products ────────│                        │                       │
```

### Token Refresh Flow

```
React                            Go Backend
  │                                 │
  │── GET /api/products ───────────>│ (access token expired)
  │<── 401 ─────────────────────────│
  │ [interceptor kicks in]          │
  │── POST /auth/refresh ──────────>│ (refresh cookie sent automatically)
  │<── { access_token } ────────────│
  │── GET /api/products (retry) ───>│
  │<── 200 products ────────────────│
```

### Session Restore (page reload)

```
React                            Go Backend
  │── GET /auth/me ────────────────>│ (no token yet → 401)
  │ [interceptor: POST /auth/refresh with cookie]
  │── POST /auth/refresh ──────────>│
  │<── { access_token } ────────────│
  │── GET /auth/me (retry) ────────>│
  │<── { user, access_token } ──────│
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/firebase` | — | Verify Firebase ID token, return JWT + set cookie |
| POST | `/auth/refresh` | Cookie | Exchange refresh token for new access token |
| POST | `/auth/logout` | — | Revoke refresh token, clear cookie |
| GET | `/auth/me` | Bearer | Get current user + fresh token |
| GET | `/api/products` | Bearer | List products (protected) |

## Setup

### Prerequisites

- Go 1.22+, Node 20+, Docker Desktop

### 1. Firebase Console

1. Create a project at [console.firebase.google.com](https://console.firebase.google.com)
2. **Authentication → Sign-in method** → enable: Email/Password, Google, GitHub
   - For GitHub: create a GitHub OAuth App at [github.com/settings/developers](https://github.com/settings/developers), set callback URL to `https://your-project.firebaseapp.com/__/auth/handler`
3. **Project settings → Service accounts → Generate new private key** (for backend `FIREBASE_CREDENTIALS_JSON`)
4. **Project settings → General → Your apps → Web** → Register app (for frontend `VITE_FIREBASE_*` values)

### 2. Start PostgreSQL

```bash
docker compose up -d postgres
```

### 3. Configure backend

```bash
cp api/.env.example api/.env
```

```env
PORT=3000
ENV=development
DATABASE_URL=postgres://postgres:postgres@localhost:5432/login_oauth?sslmode=disable
JWT_SECRET=your-32-char-random-secret
FRONTEND_URL=http://localhost:5173
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_CREDENTIALS_JSON='{"type":"service_account","project_id":"...",...}'
```

> Wrap `FIREBASE_CREDENTIALS_JSON` in **single quotes** — the JSON contains `"` characters that break `.env` parsing.

Generate JWT secret: `openssl rand -base64 32`

### 4. Configure frontend

```bash
cp web/.env.example web/.env
```

```env
VITE_API_URL=http://localhost:3000
VITE_FIREBASE_API_KEY=AIza...
VITE_FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your-project-id
```

### 5. Run

```bash
# Backend (runs migrations automatically on start)
cd api && go run ./cmd/server

# Frontend
cd web && npm install && npm run dev
```

Open [http://localhost:5173](http://localhost:5173)

## Security Notes

- **Access token**: JWT (HS256), 15 min expiry, stored in React memory only (never localStorage)
- **Refresh token**: 32-byte random, SHA-256 hashed before DB storage, 7 day expiry
- **Cookie**: `HttpOnly; SameSite=Lax` — not accessible from JavaScript, CSRF-resistant
- **CORS**: Explicit origin with `AllowCredentials: true` — wildcard `*` never used
- **Firebase**: Backend only accepts tokens it can verify — Firebase is identity, backend controls access

## Adding Providers

**Backend** — one new `case` in `api/internal/auth/firebase.go`:
```go
case "apple.com": return "apple"
```

**Frontend** — one new export in `web/src/lib/firebase.ts`:
```ts
const appleProvider = new OAuthProvider('apple.com')
export const signInWithApple = () => signInWithProvider(appleProvider)
```

No new routes, no DB changes required.
