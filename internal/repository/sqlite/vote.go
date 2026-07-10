package sqlite

type VoteRow struct {
	UserID    int64
	Target    string
	TargetID  int64
	Value     int
	CreatedAt string
	UpdatedAt string
}
