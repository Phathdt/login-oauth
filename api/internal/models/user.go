package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string    `json:"id"`
	GoogleID  string    `json:"google_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"created_at"`
}

func FindOrCreateByGoogle(db *pgxpool.Pool, googleID, email, name, picture string) (*User, error) {
	user := &User{}

	err := db.QueryRow(context.Background(),
		`SELECT id, google_id, email, name, picture, created_at FROM users WHERE google_id = $1`,
		googleID,
	).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Picture, &user.CreatedAt)

	if err == nil {
		// User exists — update name/picture in case they changed
		_, updateErr := db.Exec(context.Background(),
			`UPDATE users SET name = $1, picture = $2 WHERE google_id = $3`,
			name, picture, googleID,
		)
		if updateErr == nil {
			user.Name = name
			user.Picture = picture
		}
		return user, nil
	}

	// User does not exist — create
	err = db.QueryRow(context.Background(),
		`INSERT INTO users (google_id, email, name, picture)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (google_id) DO UPDATE SET name = EXCLUDED.name, picture = EXCLUDED.picture
		 RETURNING id, google_id, email, name, picture, created_at`,
		googleID, email, name, picture,
	).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Picture, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	return user, nil
}

func FindUserByID(db *pgxpool.Pool, id string) (*User, error) {
	user := &User{}

	err := db.QueryRow(context.Background(),
		`SELECT id, google_id, email, name, picture, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Picture, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
