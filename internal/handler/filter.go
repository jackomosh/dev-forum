package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/domain"
)

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

// FilterHandler handles filter-related HTTP requests
type FilterHandler struct {
	postRepo PostRepository
	renderer *Renderer
}

// NewFilterHandler creates a new filter handler
func NewFilterHandler(
	postRepo PostRepository,
	renderer *Renderer,
) *FilterHandler {
	return &FilterHandler{
		postRepo: postRepo,
		renderer: renderer,
	}
}

// HandleFilter applies filters to posts and returns filtered results
func (h *FilterHandler) HandleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := GetUserFromContext(r.Context())
	publicUser, _ := GetPublicUserFromContext(r.Context())

	// Parse filter parameters
	categorySlug := strings.TrimSpace(r.URL.Query().Get("category"))
	filterType := strings.TrimSpace(r.URL.Query().Get("filter"))
	sortBy := strings.TrimSpace(r.URL.Query().Get("sort"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	pageStr := r.URL.Query().Get("page")

	// Parse page number
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Build filter
	filter := domain.PostFilter{
		Kind:   domain.PostFilterAll,
		Sort:   domain.SortNewest,
		Limit:  20,
		Offset: (page - 1) * 20,
	}
	if user != nil {
		filter.ViewerID = user.ID
	}

	// Apply category filter
	if categorySlug != "" {
		categories, err := h.postRepo.GetAllCategories(r.Context())
		if err != nil {
			h.renderer.serverError(w, err)
			return
		}
		for _, cat := range categories {
			if cat.Slug == categorySlug {
				filter.Kind = domain.PostFilterCategory
				filter.CategoryID = cat.ID
				break
			}
		}
	}

	// Apply user filter
	switch filterType {
	case string(domain.PostFilterCreated):
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		filter.Kind = domain.PostFilterCreated
	case string(domain.PostFilterLiked):
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
		filter.Kind = domain.PostFilterLiked
	}

	// Apply search filter
	if search != "" {
		filter.Search = search
	}

	// Apply sort
	switch sortBy {
	case "oldest":
		filter.Sort = domain.SortOldest
	case "most_liked":
		filter.Sort = domain.SortMostLiked
	case "most_commented":
		filter.Sort = domain.SortMostCommented
	default:
		filter.Sort = domain.SortNewest
	}

	// Get filtered posts
	posts, total, err := h.postRepo.List(r.Context(), filter)
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	// Build post list items
	items := make([]PostListItem, 0, len(posts))
	for _, post := range posts {
		items = append(items, PostListItem{
			Post:     post,
			Comments: []domain.CommentWithAuthor{},
		})
	}

	// Get categories for filter UI
	categories, err := h.postRepo.GetAllCategories(r.Context())
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	// Calculate pagination
	totalPages := (total + filter.Limit - 1) / filter.Limit
	if totalPages < 1 {
		totalPages = 1
	}

	// Build pagination data
	pagination := PaginationViewData{
		Page:       page,
		Limit:      filter.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	data := PostListViewData{
		BaseViewData: BaseViewData{
			CurrentUser: publicUser,
		},
		Posts:        items,
		Categories:   categories,
		Filter:       filter,
		ActiveCat:    categorySlug,
		ActiveFilter: filterType,
	}

	// Add pagination to data
	// render the dashboard with filter applied
	h.renderer.Render(w, "dashboard.html", data)
}

// HandleGetCategories returns all available categories (AJAX)
func (h *FilterHandler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.postRepo.GetAllCategories(r.Context())
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		h.renderer.serverError(w, err)
	}
}

// HandleGetFilterOptions returns available filter options (AJAX)
func (h *FilterHandler) HandleGetFilterOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _ := GetUserFromContext(r.Context())
