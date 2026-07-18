package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"forum/internal/domain"
)

// VoteHandler handles reaction-related HTTP requests
type VoteHandler struct {
	voteRepo VoteRepository
	renderer *Renderer
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(
	voteRepo VoteRepository,
	renderer *Renderer,
) *VoteHandler {
	return &VoteHandler{
		voteRepo: voteRepo,
		renderer: renderer,
	}
}

// HandleVote processes like/dislike actions (AJAX endpoint)
func (h *VoteHandler) HandleVote(w http.ResponseWriter, r *http.Request) {
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

	// Parse JSON request
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate target type
	if req.Target != domain.VoteTargetPost {
		http.Error(w, "invalid vote target", http.StatusBadRequest)
		return
	}

	// Validate vote value
	if req.Value != domain.VoteLike && req.Value != domain.VoteDislike && req.Value != domain.VoteNone {
		http.Error(w, "invalid vote value", http.StatusBadRequest)
		return
	}

	// Create vote object
	vote := &domain.Vote{
		UserID:   user.ID,
		Target:   req.Target,
		TargetID: req.TargetID,
		Value:    req.Value,
	}

	// Save vote using repository
	if err := h.voteRepo.AddVote(r.Context(), vote); err != nil {
		log.Printf("Error saving vote: %v", err)
		http.Error(w, "Failed to save vote", http.StatusInternalServerError)
		return
	}

	// Get updated stats
	stats, err := h.voteRepo.GetPostStats(r.Context(), domain.PostID(req.TargetID))
	if err != nil {
		log.Printf("Error getting vote stats: %v", err)
		http.Error(w, "Failed to get vote stats", http.StatusInternalServerError)
		return
	}

	// Get user's current vote
	userVote, err := h.voteRepo.GetVote(r.Context(), user.ID, req.Target, req.TargetID)
	if err != nil {
		userVote = &domain.Vote{Value: domain.VoteNone}
	}

	// Prepare response
	response := VoteResponse{
		TargetID:     req.TargetID,
		LikeCount:    stats.LikeCount,
		DislikeCount: stats.DislikeCount,
		Score:        stats.Score,
		UserVote:     userVote.Value,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleGetVoteStatus gets the vote status for a post (optional endpoint)
func (h *VoteHandler) HandleGetVoteStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get post ID from query
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

	// Get current user (optional)
	user, _ := GetUserFromContext(r.Context())

	// Get vote stats
	stats, err := h.voteRepo.GetPostStats(r.Context(), domain.PostID(postID))
	if err != nil {
		log.Printf("Error getting vote stats: %v", err)
		http.Error(w, "Failed to get vote stats", http.StatusInternalServerError)
		return
	}

	// Get user's vote if logged in
	var userVoteValue domain.VoteValue = domain.VoteNone
	if user != nil {
		userVote, err := h.voteRepo.GetVote(r.Context(), user.ID, domain.VoteTargetPost, postID)
		if err == nil && userVote != nil {
			userVoteValue = userVote.Value
		}
	}

	// Prepare response
	response := VoteResponse{
		TargetID:     postID,
		LikeCount:    stats.LikeCount,
		DislikeCount: stats.DislikeCount,
		Score:        stats.Score,
		UserVote:     userVoteValue,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
