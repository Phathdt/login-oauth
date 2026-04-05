# login-oauth

Full-stack Google OAuth2 demo with JWT access tokens and HTTP-only refresh token cookies.

## Stack

| Layer     | Tech                                                                                             |
| --------- | ------------------------------------------------------------------------------------------------ |
| Frontend  | React + Vite + TypeScript, React Router v7, TanStack Query v5, Axios, Tailwind CSS v4, shadcn/ui |
| Backend   | Go + Fiber v2, golang-jwt/jwt v5, golang.org/x/oauth2                                            |
| Database  | PostgreSQL (via pgx/v5), Goose migrations                                                        |
| Dev infra | Docker Compose (postgres only)                                                                   |

## Project Structure

```
login-oauth/
├── web/                    # React frontend (port 5173)
│   └── src/
│       ├── contexts/       # Auth context (in-memory token state)
│       ├── lib/            # Axios client + interceptors, Query client
│       ├── routes/         # login, auth-callback, products
│       ├── components/     # ProtectedRoute, Navbar, ProductCard
│       └── services/       # product-service.ts
├── api/                    # Go backend (port 3000)
│   ├── cmd/server/         # Entry point
│   └── internal/
│       ├── auth/           # JWT, OAuth config, middleware
│       ├── handlers/       # oauth, auth, product handlers
│       ├── models/         # user, refresh_token, product
│       ├── config/         # Env config loader
│       └── database/       # pgx pool + Goose migrations
├── docker-compose.yml      # PostgreSQL only
└── README.md
```

## Login Flow

```
Browser                    Frontend (5173)           Backend (3000)           Google
  │                             │                         │                      │
  │── visit /login ────────────>│                         │                      │
  │                             │                         │                      │
  │── click "Sign in" ─────────>│                         │                      │
  │                             │── redirect ────────────>│                      │
  │                             │  GET /auth/google/login  │                      │
  │                             │                         │── redirect ─────────>│
  │                             │                         │  (with state param)  │
  │                             │                         │                      │
  │<──────────────────────────────────────────────────── Google consent page ────│
  │── user approves ──────────────────────────────────────────────────────────── │
  │                             │                         │<─ callback w/ code ──│
  │                             │                         │                      │
  │                             │                         │ 1. validate state    │
  │                             │                         │ 2. exchange code     │
  │                             │                         │ 3. fetch userinfo    │
  │                             │                         │ 4. upsert user in DB │
  │                             │                         │ 5. generate JWT (15m)│
  │                             │                         │ 6. generate refresh  │
  │                             │                         │    token (7d, DB)    │
  │                             │                         │ 7. set HTTP-only     │
  │                             │                         │    cookie            │
  │                             │<── redirect ────────────│                      │
  │                             │  /auth/callback          │                      │
  │                             │  ?access_token=<jwt>     │                      │
  │                             │                         │                      │
  │                             │ store token in memory   │                      │
  │                             │── GET /auth/me ─────────>│                      │
  │                             │  Authorization: Bearer  │                      │
  │                             │<── user info ───────────│                      │
  │<── redirect to /products ───│                         │                      │
```

### Token Refresh Flow

```
Frontend                         Backend
  │                                 │
  │── GET /api/products ───────────>│ (access token expired)
  │<── 401 Unauthorized ────────────│
  │                                 │
  │ [interceptor kicks in]          │
  │── POST /auth/refresh ──────────>│ (sends refresh_token cookie automatically)
  │<── { access_token: "..." } ─────│ (new JWT, 15min)
  │                                 │
  │── GET /api/products ───────────>│ (retry with new token)
  │<── 200 [ products ] ────────────│
```

### Session Restore (page reload)

```
Frontend                         Backend
  │                                 │
  │── GET /auth/me ────────────────>│ (no token yet)
  │<── 401 ─────────────────────────│
  │ [interceptor: POST /auth/refresh with cookie]
  │── POST /auth/refresh ──────────>│
  │<── { access_token } ────────────│
  │── GET /auth/me (retry) ────────>│
  │<── { user, access_token } ──────│
  │ restore auth state in memory    │
```

## API Endpoints

| Method | Path                    | Auth   | Description                                         |
| ------ | ----------------------- | ------ | --------------------------------------------------- |
| GET    | `/auth/google/login`    | —      | Redirect to Google consent                          |
| GET    | `/auth/google/callback` | —      | Handle OAuth code, set cookie, redirect to frontend |
| POST   | `/auth/refresh`         | Cookie | Exchange refresh token for new access token         |
| POST   | `/auth/logout`          | —      | Revoke refresh token, clear cookie                  |
| GET    | `/auth/me`              | Bearer | Get current user info + fresh token                 |
| GET    | `/api/products`         | Bearer | List products (protected)                           |

## Setup

### Prerequisites

- Go 1.22+
- Node 20+
- Docker Desktop

### 1. Start PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Configure backend

```bash
cp api/.env.example api/.env
```

Edit `api/.env`:

```env
PORT=3000
ENV=development
DATABASE_URL=postgres://postgres:postgres@localhost:5432/login_oauth?sslmode=disable
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
JWT_SECRET=your-32-char-random-secret
FRONTEND_URL=http://localhost:5173
```

Generate a JWT secret:

```bash
openssl rand -base64 32
```

### 3. Google OAuth credentials

1. Go to [console.cloud.google.com](https://console.cloud.google.com)
2. Create project → **APIs & Services** → **Credentials** → **OAuth 2.0 Client ID**
3. Application type: **Web application**
4. Authorized redirect URI: `http://localhost:3000/auth/google/callback`
5. Copy Client ID and Secret into `api/.env`

### 4. Run backend (migrations run automatically on start)

```bash
cd api && go run ./cmd/server
```

### 5. Run frontend

```bash
cd web && npm install && npm run dev
```

Open `http://localhost:5173`

## Security Notes

- **Access token**: JWT (HS256), 15 min expiry, stored in React memory only (never localStorage)
- **Refresh token**: 32-byte random, SHA256-hashed before DB storage, 7 day expiry
- **Cookie**: `HttpOnly; SameSite=Lax` — not accessible from JavaScript, CSRF-resistant
- **CORS**: Explicit origin (`http://localhost:5173`) with `AllowCredentials: true` — `*` is never used
- **State param**: CSRF protection on OAuth callback (validated before code exchange)
