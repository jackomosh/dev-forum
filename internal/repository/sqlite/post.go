package sqlite

type PostRow struct {
	ID        int64
	AuthorID  int64
	Title     string
	Body      string
	Status    string
	CreatedAt string
	UpdatedAt string
}

type CategoryRow struct {
	ID          int64
	Name        string
	Slug        string
	Description string
	CreatedAt   string
}

type PostCategoryRow struct {
	PostID     int64
	CategoryID int64
}

type PostStatsRow struct {
	PostID   int64
	Comments int
	Likes    int
	Dislikes int
	Score    int
	UserVote int
}
