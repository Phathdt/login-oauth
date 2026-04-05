-- name: FindOrCreateUser :one
INSERT INTO users (firebase_uid, provider, email, name, picture)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (firebase_uid) DO UPDATE
  SET provider = EXCLUDED.provider,
      email = EXCLUDED.email,
      name = EXCLUDED.name,
      picture = EXCLUDED.picture
RETURNING *;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;
