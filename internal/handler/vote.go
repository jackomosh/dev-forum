package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"forum/internal/domain"
	"forum/internal/repository"
)

// VoteHandler handles reaction-related HTTP requests
type VoteHandler struct {
	voteRepo repository.VoteRepository
	postRepo repository.PostRepository
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(
	voteRepo repository.VoteRepository,
	postRepo repository.PostRepository,
) *VoteHandler {
	return &VoteHandler{
		voteRepo: voteRepo,
		postRepo: postRepo,
	}
}

// HandleVote processes like/dislike actions (AJAX endpoint)
func (h *VoteHandler) HandleVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse JSON request
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate target type - only posts for now (can extend to comments later)
	if req.Target != domain.VoteTargetPost {
		http.Error(w, "Invalid vote target", http.StatusBadRequest)
		return
	}

	// Validate vote value
	if req.Value != domain.VoteUp && req.Value != domain.VoteDown && req.Value != domain.VoteNone {
		http.Error(w, "Invalid vote value", http.StatusBadRequest)
		return
	}

	// Convert VoteValue to string for repository
	voteType := ""
	switch req.Value {
	case domain.VoteUp:
		voteType = "like"
	case domain.VoteDown:
		voteType = "dislike"
	case domain.VoteNone:
		voteType = "none"
	}

	// Save vote using repository
	if err := h.voteRepo.SaveVote(user.ID, req.TargetID, voteType); err != nil {
		log.Printf("Error saving vote: %v", err)
		http.Error(w, "Failed to save vote", http.StatusInternalServerError)
		return
	}

	// Get updated counts
	likes, dislikes, err := h.voteRepo.GetVoteCounts(req.TargetID)
	if err != nil {
		log.Printf("Error getting vote counts: %v", err)
		http.Error(w, "Failed to get vote counts", http.StatusInternalServerError)
		return
	}

	// Get user's current vote
	userVote, err := h.voteRepo.GetUserVote(user.ID, req.TargetID)
	if err != nil {
		// Default to "none" if no vote found
		userVote = &domain.Vote{VoteType: "none"}
	}

	// Convert to VoteValue
	var userVoteValue domain.VoteValue
	switch userVote.VoteType {
	case "like":
		userVoteValue = domain.VoteUp
	case "dislike":
		userVoteValue = domain.VoteDown
	default:
		userVoteValue = domain.VoteNone
	}

	// Prepare response using your VoteResponse struct
	response := VoteResponse{
		TargetID:     req.TargetID,
		LikeCount:    likes,
		DislikeCount: dislikes,
		Score:        likes - dislikes,
		UserVote:     userVoteValue,
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get post ID from query
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get current user (optional)
	user, _ := GetUserFromContext(r.Context())

	// Get vote counts
	likes, dislikes, err := h.voteRepo.GetVoteCounts(postID)
	if err != nil {
		log.Printf("Error getting vote counts: %v", err)
		http.Error(w, "Failed to get vote counts", http.StatusInternalServerError)
		return
	}

	// Get user's vote if logged in
	var userVoteValue domain.VoteValue = domain.VoteNone
	if user != nil {
		userVote, err := h.voteRepo.GetUserVote(user.ID, postID)
		if err == nil && userVote != nil {
			switch userVote.VoteType {
			case "like":
				userVoteValue = domain.VoteUp
			case "dislike":
				userVoteValue = domain.VoteDown
			}
		}
	}

	// Prepare response
	response := VoteResponse{
		TargetID:     postID,
		LikeCount:    likes,
		DislikeCount: dislikes,
		Score:        likes - dislikes,
		UserVote:     userVoteValue,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}