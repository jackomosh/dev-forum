package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

	comment := &domain.Comment{
		PostID:   domain.PostID(postID),
		AuthorID: user.ID,
		Body:     body,
		Status:   domain.CommentStatusVisible,
	}

	if err := h.commentRepo.Create(r.Context(), comment); err != nil {
		h.renderer.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/post/"+strconv.FormatInt(postID, 10), http.StatusSeeOther)
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
