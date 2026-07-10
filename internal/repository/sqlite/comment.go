package sqlite

type CommentRow struct {
	ID        int64
	PostID    int64
	AuthorID  int64
	Body      string
	Status    string
	CreatedAt string
	UpdatedAt string
}

type CommentStatsRow struct {
	CommentID int64
	Likes     int
	Dislikes  int
	Score     int
}
