package handler

import (
	"context"

	"forum/internal/domain"
)

// ============================================
// BASE TYPES
// ============================================

// BaseViewData contains common data for all pages
type BaseViewData struct {
	CurrentUser *domain.PublicUser
	Error       string
	Success     string
}

// ============================================
// VIEW STRUCTS
// ============================================

// PostListItem represents a single post in the feed with its comments
type PostListItem struct {
	Post     domain.PostWithAuthor
	Comments []domain.CommentWithAuthor
}

// PostListViewData contains data for the post feed page
type PostListViewData struct {
	BaseViewData
	Posts        []PostListItem
	Categories   []domain.Category
	Filter       domain.PostFilter
	ActiveCat    string
	ActiveFilter string
}

// PostDetailViewData contains data for the single post page
type PostDetailViewData struct {
	BaseViewData
	Post        domain.PostWithAuthor
	Comments    []domain.CommentWithAuthor
	CommentForm CommentForm
}

// CommentViewData contains data for a single comment view
type CommentViewData struct {
	BaseViewData
	Comment domain.CommentWithAuthor
}

// CommentForm represents the form data for creating a comment
type CommentForm struct {
	PostID domain.PostID
	Body   string
}

// RegisterRequest represents registration form data
type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

// AuthViewData contains data for auth page rendering
type AuthViewData struct {
	BaseViewData
	Form RegisterRequest
}

// VoteRequest represents the vote data from client
type VoteRequest struct {
	Target   domain.VoteTarget `json:"target"`
	TargetID int64             `json:"target_id"`
	Value    domain.VoteValue  `json:"value"`
}

// VoteResponse represents the vote response to client
type VoteResponse struct {
	TargetID     int64            `json:"target_id"`
	LikeCount    int              `json:"like_count"`
	DislikeCount int              `json:"dislike_count"`
	Score        int              `json:"score"`
	UserVote     domain.VoteValue `json:"user_vote"`
}

// FilterForm represents the filter form data
type FilterForm struct {
	CategoryID domain.CategoryID
	Kind       domain.PostFilterKind
	Search     string
	Sort       domain.SortOrder
	Page       int
}

// PaginationViewData contains pagination data
type PaginationViewData struct {
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

// PostForm represents the form data for creating/editing a post
type PostForm struct {
	Title       string
	Body        string
	CategoryIDs []domain.CategoryID
}

// ============================================
// SERVICE INTERFACE
// ============================================

// AuthService defines the authentication business logic
type AuthService interface {
	RegisterUser(ctx context.Context, username, email, password string) (*domain.User, error)
	AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error)
	CreateSession(ctx context.Context, userID domain.UserID) (string, error)
	ValidateSession(ctx context.Context, token string) (*domain.User, error)
	DeleteSession(ctx context.Context, token string) error
}
