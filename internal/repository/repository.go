package repository

import "forum/internal/domain"
// RECORD STRUCTS - Data Transfer Objects (DTOs)
// These are used to transfer data between layers

// UserRecord represents a user data transfer object
type UserRecord struct {
	User domain.User
}

// SessionRecord represents a session with user data
type SessionRecord struct {
	Session domain.Session
	User    domain.User
}

// PostRecord represents a post with all its related data
type PostRecord struct {
	Post       domain.Post
	Author     domain.User
	Categories []domain.Category
	Stats      domain.PostStats
	UserVote   domain.VoteValue
}

// CommentRecord represents a comment with all its related data
type CommentRecord struct {
	Comment  domain.Comment
	Author   domain.User
	Stats    domain.CommentStats
	UserVote domain.VoteValue
}

// QUERY STRUCTS - For filtering and pagination

// PostQuery represents a post query with filters
type PostQuery struct {
	Filter   domain.PostFilter
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

// CommentQuery represents a comment query with pagination
type CommentQuery struct {
	PostID domain.PostID
	Limit  int
	Offset int
}

// VoteQuery represents a vote query
type VoteQuery struct {
	UserID   domain.UserID
	Target   domain.VoteTarget
	TargetID int64
}

// RESULT STRUCTS - For query results

// PostListResult represents a paginated list of posts
type PostListResult struct {
	Posts      []PostRecord
	TotalCount int
	Page       int
	PageSize   int
}

// CommentListResult represents a paginated list of comments
type CommentListResult struct {
	Comments   []CommentRecord
	TotalCount int
	Page       int
	PageSize   int
}

// VoteResult represents a vote operation result
type VoteResult struct {
	Likes    int
	Dislikes int
	UserVote domain.VoteValue
}

// UserRepository defines database operations for users
type UserRepository interface {
	// Create a new user
	CreateUser(user *domain.User) error
	
	// Get user by various identifiers
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByID(id int64) (*domain.User, error)
	
	// Update user information
	UpdateUser(user *domain.User) error
	
	// Delete a user (soft delete or hard delete)
	DeleteUser(id int64) error
	
	// Check if user exists
	UserExists(email string) (bool, error)
}

// SessionRepository defines database operations for sessions
type SessionRepository interface {
	// Create a new session
	CreateSession(session *domain.Session) error
	
	// Get a session by token
	GetSessionByToken(token string) (*domain.Session, error)
	
	// Get all sessions for a user
	GetSessionsByUserID(userID int64) ([]*domain.Session, error)
	
	// Delete a specific session
	DeleteSession(token string) error
	
	// Delete all sessions for a user
	DeleteSessionsByUserID(userID int64) error
	
	// Delete expired sessions
	DeleteExpiredSessions() error
	
	// Extend session expiration
	ExtendSession(token string, duration int) error
}

// PostRepository defines database operations for posts
type PostRepository interface {
	// Create a new post
	CreatePost(post *domain.Post, categories []string) error
	
	// Get a single post by ID
	GetPostByID(id int64) (*domain.Post, error)
	
	// Get posts with filtering and pagination
	GetPosts(query PostQuery) (PostListResult, error)
	
	// Get posts created by a specific user
	GetPostsByUserID(userID int64, limit, offset int) ([]*domain.Post, error)
	
	// Get posts liked by a specific user
	GetPostsLikedByUser(userID int64, limit, offset int) ([]*domain.Post, error)
	
	// Update a post
	UpdatePost(post *domain.Post) error
	
	// Delete a post
	DeletePost(id int64) error
	
	// Get all categories
	GetAllCategories() ([]*domain.Category, error)
	
	// Create a new category
	CreateCategory(name string) (*domain.Category, error)
	
	// Get category by ID
	GetCategoryByID(id int64) (*domain.Category, error)
	
	// Get category by name
	GetCategoryByName(name string) (*domain.Category, error)
}

// CommentRepository defines database operations for comments
type CommentRepository interface {
	// Create a new comment
	CreateComment(comment *domain.Comment) error
	
	// Get a comment by ID
	GetCommentByID(id int64) (*domain.Comment, error)
	
	// Get comments for a post with pagination
	GetCommentsByPostID(query CommentQuery) (CommentListResult, error)
	
	// Get comments by a user
	GetCommentsByUserID(userID int64, limit, offset int) ([]*domain.Comment, error)
	
	// Update a comment
	UpdateComment(comment *domain.Comment) error
	
	// Delete a comment
	DeleteComment(id int64) error
	
	// Get comment count for a post
	GetCommentCount(postID int64) (int, error)
}

// VoteRepository defines database operations for votes
type VoteRepository interface {
	// Save a vote (insert or update)
	SaveVote(userID, postID int64, voteType string) error
	
	// Get vote counts for a post
	GetVoteCounts(postID int64) (likes, dislikes int, err error)
	
	// Get a user's vote on a post
	GetUserVote(userID, postID int64) (*domain.Vote, error)
	
	// Get all votes for a user
	GetUserVotes(userID int64) ([]*domain.Vote, error)
	
	// Delete a vote
	DeleteVote(userID, postID int64) error
	
	// Get users who liked a post
	GetUsersWhoLiked(postID int64) ([]*domain.User, error)
	
	// Get users who disliked a post
	GetUsersWhoDisliked(postID int64) ([]*domain.User, error)
}


// RepositoryFactory defines the factory for creating repositories
type RepositoryFactory interface {
	// Create user repository
	NewUserRepository() UserRepository
	
	// Create session repository
	NewSessionRepository() SessionRepository
	
	// Create post repository
	NewPostRepository() PostRepository
	
	// Create comment repository
	NewCommentRepository() CommentRepository
	
	// Create vote repository
	NewVoteRepository() VoteRepository
}

// RepositoryError represents a repository error with context
type RepositoryError struct {
	Code    string
	Message string
	Err     error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// Error codes
const (
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeDuplicate     = "DUPLICATE"
	ErrCodeForeignKey    = "FOREIGN_KEY"
	ErrCodeConstraint    = "CONSTRAINT"
	ErrCodeInvalid       = "INVALID"
	ErrCodeInternal      = "INTERNAL"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
)

// Predefined errors
var (
	// User errors
	ErrUserNotFound      = &RepositoryError{Code: ErrCodeNotFound, Message: "user not found"}
	ErrDuplicateEmail    = &RepositoryError{Code: ErrCodeDuplicate, Message: "email already exists"}
	ErrDuplicateUsername = &RepositoryError{Code: ErrCodeDuplicate, Message: "username already exists"}
	
	// Session errors
	ErrSessionNotFound   = &RepositoryError{Code: ErrCodeNotFound, Message: "session not found"}
	ErrSessionExpired    = &RepositoryError{Code: ErrCodeInvalid, Message: "session expired"}
	
	// Post errors
	ErrPostNotFound      = &RepositoryError{Code: ErrCodeNotFound, Message: "post not found"}
	ErrCategoryNotFound  = &RepositoryError{Code: ErrCodeNotFound, Message: "category not found"}
	ErrCategoryExists    = &RepositoryError{Code: ErrCodeDuplicate, Message: "category already exists"}
	
	// Comment errors
	ErrCommentNotFound   = &RepositoryError{Code: ErrCodeNotFound, Message: "comment not found"}
	
	// Vote errors
	ErrVoteNotFound      = &RepositoryError{Code: ErrCodeNotFound, Message: "vote not found"}
	ErrInvalidVoteType   = &RepositoryError{Code: ErrCodeInvalid, Message: "invalid vote type"}
	
	// Generic errors
	ErrDatabaseError     = &RepositoryError{Code: ErrCodeInternal, Message: "database error"}
	ErrInvalidInput      = &RepositoryError{Code: ErrCodeInvalid, Message: "invalid input"}
)

// Helper function to wrap errors
func WrapError(err error, message string) *RepositoryError {
	if err == nil {
		return nil
	}
	return &RepositoryError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
	}
}