package handler

import (
	"net/http"
	"strconv"
	"strings"

	"forum/internal/domain"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postRepo    PostRepository
	commentRepo CommentRepository
	voteRepo    VoteRepository
	renderer    *Renderer
}

// NewPostHandler creates a new posts handler
func NewPostHandler(
	postRepo PostRepository,
	commentRepo CommentRepository,
	voteRepo VoteRepository,
	renderer *Renderer,
) *PostHandler {
	return &PostHandler{
		postRepo:    postRepo,
		commentRepo: commentRepo,
		voteRepo:    voteRepo,
		renderer:    renderer,
	}
}

// HandleHomePage displays the main feed with posts
func (h *PostHandler) HandleHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	user, _ := GetPublicUserFromContext(r.Context())

	data := BaseViewData{
		CurrentUser: user,
	}
	h.renderer.Render(w, "index.html", data)
}

// HandleDashboard displays the dashboard with posts
func (h *PostHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	publicUser, _ := GetPublicUserFromContext(r.Context())

	categories, err := h.postRepo.GetAllCategories(r.Context())
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	activeCategory := strings.TrimSpace(r.URL.Query().Get("category"))
	activeFilter := strings.TrimSpace(r.URL.Query().Get("filter"))

	filter := domain.PostFilter{
		Kind:   domain.PostFilterAll,
		Sort:   domain.SortNewest,
		Limit:  20,
		Offset: 0,
	}
	filter.ViewerID = currentUser.ID

	if activeCategory != "" {
		for _, cat := range categories {
			if cat.Slug == activeCategory {
				filter.Kind = domain.PostFilterCategory
				filter.CategoryID = cat.ID
				break
			}
		}
	}

	switch activeFilter {
	case "created":
		filter.Kind = domain.PostFilterCreated
		filter.AuthorID = currentUser.ID
	case "liked":
		filter.Kind = domain.PostFilterLiked
		filter.ViewerID = currentUser.ID
	default:
		activeFilter = ""
	}

	// Get posts using List method
	posts, _, err := h.postRepo.List(r.Context(), filter)
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	// Build post list items
	items := make([]PostListItem, 0, len(posts))
	for _, post := range posts {
		// Get comments for this post
		comments, _, err := h.commentRepo.GetByPostID(r.Context(), post.Post.ID, 5, 0)
		if err != nil {
			comments = []domain.CommentWithAuthor{}
		}
		items = append(items, PostListItem{
			Post:     post,
			Comments: comments,
		})
	}

	data := PostListViewData{
		BaseViewData: BaseViewData{
			CurrentUser: publicUser,
		},
		Posts:        items,
		Categories:   categories,
		Filter:       filter,
		ActiveCat:    activeCategory,
		ActiveFilter: activeFilter,
	}

	h.renderer.Render(w, "dashboard.html", data)
}

// HandleViewPost displays a single post with comments
func (h *PostHandler) HandleViewPost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/post/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.postRepo.GetByID(r.Context(), domain.PostID(id))
	if err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	currentUser, _ := GetPublicUserFromContext(r.Context())

	comments, _, err := h.commentRepo.GetByPostID(r.Context(), domain.PostID(id), 100, 0)
	if err != nil {
		comments = []domain.CommentWithAuthor{}
	}

	data := PostDetailViewData{
		BaseViewData: BaseViewData{
			CurrentUser: currentUser,
		},
		Post:     *post,
		Comments: comments,
		CommentForm: CommentForm{
			PostID: domain.PostID(id),
		},
	}

	h.renderer.Render(w, "post_detail.html", data)
}

// HandleCreatePostPage displays the create post form
func (h *PostHandler) HandleCreatePostPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	publicUser, _ := GetPublicUserFromContext(r.Context())
	categories, err := h.postRepo.GetAllCategories(r.Context())
	if err != nil {
		h.renderer.serverError(w, err)
		return
	}

	data := struct {
		BaseViewData
		Categories []domain.Category
	}{
		BaseViewData: BaseViewData{
			CurrentUser: publicUser,
		},
		Categories: categories,
	}

	h.renderer.Render(w, "create_post.html", data)
}

// HandleCreatePost processes post creation
func (h *PostHandler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	body := strings.TrimSpace(r.FormValue("body"))
	categories := r.Form["categories"]

	if title == "" || body == "" {
		http.Error(w, "title and body are required", http.StatusBadRequest)
		return
	}

	// Parse category IDs
	var categoryIDs []domain.CategoryID
	for _, catName := range categories {
		// Try to get category by name
		cat, err := h.postRepo.GetCategoryByName(r.Context(), catName)
		if err != nil {
			// Create category if it doesn't exist
			cat, err = h.postRepo.CreateCategory(r.Context(), catName)
			if err != nil {
				h.renderer.serverError(w, err)
				return
			}
		}
		if cat != nil {
			categoryIDs = append(categoryIDs, cat.ID)
		}
	}

	post := &domain.Post{
		Title:    title,
		Body:     body,
		AuthorID: user.ID,
		Status:   domain.PostStatusPublished,
	}

	if err := h.postRepo.Create(r.Context(), post, categoryIDs); err != nil {
		h.renderer.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
