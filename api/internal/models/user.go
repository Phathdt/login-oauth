package models

// User is kept as a domain alias used by auth handlers for constructing JWT claims
// without a database round-trip. The authoritative DB struct is internal/db.User.
type User struct {
	ID      string
	Email   string
	Name    string
	Picture string
}
