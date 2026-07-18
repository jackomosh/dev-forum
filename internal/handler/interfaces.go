package handler

import (
	"context"

	"forum/internal/domain"
)

// ============================================
// REPOSITORY INTERFACES
// ============================================

// UserRepository defines database operations for users
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// SessionRepository defines database operations for sessions
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID domain.UserID) ([]domain.Session, error)
	Delete(ctx context.Context, id domain.SessionID) error
	DeleteByUserID(ctx context.Context, userID domain.UserID) error
	CleanupExpired(ctx context.Context) error
}

// PostRepository defines database operations for posts
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post, categoryIDs []domain.CategoryID) error
	GetByID(ctx context.Context, id domain.PostID) (*domain.PostWithAuthor, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id domain.PostID) error
	List(ctx context.Context, filter domain.PostFilter) ([]domain.PostWithAuthor, int, error)
	GetCategoriesByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error)
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
	GetCategoryByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*domain.Category, error)
	CreateCategory(ctx context.Context, name string) (*domain.Category, error)
}

// CommentRepository defines database operations for comments
type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, id domain.CommentID) (*domain.CommentWithAuthor, error)
	GetByPostID(ctx context.Context, postID domain.PostID, limit, offset int) ([]domain.CommentWithAuthor, int, error)
	Update(ctx context.Context, comment *domain.Comment) error
	Delete(ctx context.Context, id domain.CommentID) error
}

// CategoryRepository defines database operations for categories
type CategoryRepository interface {
	GetAll(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error)
	GetByPostID(ctx context.Context, postID domain.PostID) ([]domain.Category, error)
	Create(ctx context.Context, category *domain.Category) error
}

// VoteRepository defines database operations for votes
type VoteRepository interface {
	AddVote(ctx context.Context, vote *domain.Vote) error
	GetVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) (*domain.Vote, error)
	GetPostStats(ctx context.Context, postID domain.PostID) (*domain.PostStats, error)
	GetCommentStats(ctx context.Context, commentID domain.CommentID) (*domain.CommentStats, error)
	RemoveVote(ctx context.Context, userID domain.UserID, target domain.VoteTarget, targetID int64) error
	GetUserVotedPostIDs(ctx context.Context, userID domain.UserID) ([]domain.PostID, error)
}
