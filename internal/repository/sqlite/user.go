package sqlite

type UserRow struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    string
	UpdatedAt    string
}

type SessionRow struct {
	ID        string
	UserID    int64
	ExpiresAt string
	CreatedAt string
}
