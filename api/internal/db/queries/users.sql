-- name: FindOrCreateUser :one
INSERT INTO users (google_id, email, name, picture)
VALUES ($1, $2, $3, $4)
ON CONFLICT (google_id) DO UPDATE
  SET email = EXCLUDED.email,
      name = EXCLUDED.name,
      picture = EXCLUDED.picture
RETURNING *;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;
