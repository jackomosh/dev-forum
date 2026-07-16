package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/domain"
)

// CommentHandler handles comment-related HTTP requests
type CommentHandler struct {
	commentRepo CommentRepository
	renderer    *Renderer
}

// NewCommentHandler creates a new comments handler
func NewCommentHandler(
	commentRepo CommentRepository,
	renderer *Renderer,
) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
		renderer:    renderer,
	}
}

// HandleCreateComment processes comment creation
func (h *CommentHandler) HandleCreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(r.FormValue("post_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(r.FormValue("comment_body"))
	if body == "" {
		http.Error(w, "comment body is required", http.StatusBadRequest)
		return
	}

	// Create comment using your CommentForm struct
	comment := &domain.Comment{
		PostID:    domain.PostID(postID),
		AuthorID:  user.ID,
		Body:      body,
		Status:    domain.CommentStatusVisible,
		CreatedAt: time.Now(),
	}

	// Save comment using repository
	if err := h.commentRepo.Create(r.Context(), comment); err != nil {
		h.renderer.serverError(w, err)
		return
	}

	// Redirect back to post
	http.Redirect(w, r, "/post/"+strconv.FormatInt(postID, 10), http.StatusSeeOther)
}

// HandleViewComment displays a single comment
func (h *CommentHandler) HandleViewComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract comment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/comment/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid comment ID", http.StatusBadRequest)
		return
	}

	comment, err := h.commentRepo.GetByID(r.Context(), domain.CommentID(id))
	if err != nil {
		http.Error(w, "comment not found", http.StatusNotFound)
		return
	}

	user, _ := GetPublicUserFromContext(r.Context())

	// Use your CommentViewData struct
	data := CommentViewData{
		BaseViewData: BaseViewData{
			CurrentUser: user,
		},
		Comment: *comment,
	}

	h.renderer.Render(w, "comment_detail.html", data)
}

// HandleDeleteComment deletes a comment
func (h *CommentHandler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseInt(r.FormValue("comment_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid comment ID", http.StatusBadRequest)
		return
	}

	// Get comment to check ownership
	comment, err := h.commentRepo.GetByID(r.Context(), domain.CommentID(commentID))
	if err != nil {
		http.Error(w, "comment not found", http.StatusNotFound)
		return
	}

	// Check if user is the author or admin
	if comment.AuthorID != user.ID && user.Role != domain.UserRoleAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Delete comment
	if err := h.commentRepo.Delete(r.Context(), domain.CommentID(commentID)); err != nil {
		h.renderer.serverError(w, err)
		return
	}

	// Redirect back to post
	http.Redirect(w, r, "/post/"+strconv.FormatInt(int64(comment.PostID), 10), http.StatusSeeOther)
}

// HandleGetComments retrieves comments for a post (AJAX endpoint)
func (h *CommentHandler) HandleGetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "post_id required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	comments, total, err := h.commentRepo.GetByPostID(r.Context(), domain.PostID(postID), limit, offset)
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	// Prepare response
	response := struct {
		Comments []domain.CommentWithAuthor `json:"comments"`
		Total    int                        `json:"total"`
		Limit    int                        `json:"limit"`
		Offset   int                        `json:"offset"`
	}{
		Comments: comments,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.renderer.serverError(w, err)
	}
}