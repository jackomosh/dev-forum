package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/domain"
)

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

	// Apply user filter - using AuthorID for created posts
	switch filterType {
	case string(domain.PostFilterCreated):
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		filter.Kind = domain.PostFilterCreated
		filter.AuthorID = user.ID
	case string(domain.PostFilterLiked):
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		filter.Kind = domain.PostFilterLiked
		filter.ViewerID = user.ID
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
		filter.Sort = domain.SortMostComment
	default:
		filter.Sort = domain.SortNewest
	}

	// Get filtered posts
	posts, _, err := h.postRepo.List(r.Context(), filter)
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

	// Build filter options
	options := struct {
		Categories   []domain.Category `json:"categories"`
		SortOptions  []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"sort_options"`
		FilterOptions []struct {
			Value        string `json:"value"`
			Label        string `json:"label"`
			AuthRequired bool   `json:"auth_required"`
		} `json:"filter_options"`
	}{
		SortOptions: []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		}{
			{Value: "newest", Label: "Newest First"},
			{Value: "oldest", Label: "Oldest First"},
			{Value: "most_liked", Label: "Most Liked"},
			{Value: "most_commented", Label: "Most Commented"},
		},
		FilterOptions: []struct {
			Value        string `json:"value"`
			Label        string `json:"label"`
			AuthRequired bool   `json:"auth_required"`
		}{
			{Value: "", Label: "All Posts", AuthRequired: false},
			{Value: "created", Label: "My Posts", AuthRequired: true},
			{Value: "liked", Label: "Liked Posts", AuthRequired: true},
		},
	}

	// Get categories
	categories, err := h.postRepo.GetAllCategories(r.Context())
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}
	options.Categories = categories

	// If user is not logged in, remove auth-required options
	if user == nil {
		var publicFilterOptions []struct {
			Value        string `json:"value"`
			Label        string `json:"label"`
			AuthRequired bool   `json:"auth_required"`
		}
		for _, opt := range options.FilterOptions {
			if !opt.AuthRequired {
				publicFilterOptions = append(publicFilterOptions, opt)
			}
		}
		options.FilterOptions = publicFilterOptions
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(options); err != nil {
		h.renderer.serverError(w, err)
	}
}
