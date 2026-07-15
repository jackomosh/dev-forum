package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/domain"
	"forum/internal/repository"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	voteRepo    repository.VoteRepository
	templates   *template.Template
}

// NewPostHandler creates a new posts handler
func NewPostHandler(
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	voteRepo repository.VoteRepository,
	templates *template.Template,
) *PostHandler {
	return &PostHandler{
		postRepo:    postRepo,
		commentRepo: commentRepo,
		voteRepo:    voteRepo,
		templates:   templates,
	}
}

// HandleHomePage displays the main feed with posts
func (h *PostHandler) HandleHomePage(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	user, _ := GetUserFromContext(r.Context())

	// Parse filter parameters
	filter := r.URL.Query().Get("filter")
	category := r.URL.Query().Get("category")

	// Build filter
	postFilter := &domain.PostFilter{
		Kind: domain.FilterAll,
	}

	if category != "" {
		postFilter.Kind = domain.FilterCategory
		postFilter.Category = category
	}

	if user != nil {
		if filter == "created" {
			postFilter.Kind = domain.FilterCreated
			postFilter.UserID = user.ID
		} else if filter == "liked" {
			postFilter.Kind = domain.FilterLiked
			postFilter.UserID = user.ID
		}
	}

	// Fetch posts using repository
	posts, err := h.postRepo.GetPosts(postFilter)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		posts = []*domain.Post{}
	}

	// Fetch all categories for the filter sidebar
	categories, err := h.postRepo.GetAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		categories = []*domain.Category{}
	}

	// Convert to PostListViewData
	var publicUser *domain.PublicUser
	if user != nil {
		publicUser = &domain.PublicUser{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.String(),
		}
	}

	// Build PostListItems
	var postItems []PostListItem
	for _, post := range posts {
		// Get comments for this post
		comments, _ := h.commentRepo.GetCommentsByPostID(post.ID)
		
		var commentAuthors []domain.CommentWithAuthor
		for _, comment := range comments {
			// Get author for each comment
			author, _ := h.userRepo.GetUserByID(comment.UserID)
			commentAuthors = append(commentAuthors, domain.CommentWithAuthor{
				Comment: *comment,
				Author:  domain.PublicUser{
					ID:        author.ID,
					Username:  author.Username,
					CreatedAt: author.CreatedAt.String(),
				},
			})
		}

		// Get author for post
		author, _ := h.userRepo.GetUserByID(post.UserID)
		
		postItems = append(postItems, PostListItem{
			Post: domain.PostWithAuthor{
				Post:   *post,
				Author: domain.PublicUser{
					ID:        author.ID,
					Username:  author.Username,
					CreatedAt: author.CreatedAt.String(),
				},
			},
			Comments: commentAuthors,
		})
	}

	data := PostListViewData{
		BaseViewData: BaseViewData{
			CurrentUser: publicUser,
			Error:       "",
		},
		Posts:        postItems,
		Categories:   categories,
		Filter:       *postFilter,
		ActiveCat:    category,
		ActiveFilter: filter,
	}

	// Render template
	templateName := "index.html"
	if h.templates.Lookup("dashboard.html") != nil {
		templateName = "dashboard.html"
	}
	h.renderTemplate(w, templateName, data)
}

// HandleViewPost displays a single post with comments
func (h *PostHandler) HandleViewPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	path := r.URL.Path
	idStr := ""

	if strings.HasPrefix(path, "/post/") {
		idStr = strings.TrimPrefix(path, "/post/")
	} else {
		idStr = r.URL.Query().Get("id")
	}

	if idStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get post using repository
	post, err := h.postRepo.GetPostByID(id)
	if err != nil {
		log.Printf("Error fetching post: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Get author
	author, err := h.userRepo.GetUserByID(post.UserID)
	if err != nil {
		log.Printf("Error fetching author: %v", err)
		author = &domain.User{Username: "Unknown"}
	}

	// Get comments
	comments, err := h.commentRepo.GetCommentsByPostID(id)
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		comments = []*domain.Comment{}
	}

	// Build comment list with authors
	var commentAuthors []domain.CommentWithAuthor
	for _, comment := range comments {
		commentAuthor, err := h.userRepo.GetUserByID(comment.UserID)
		if err != nil {
			commentAuthor = &domain.User{Username: "Unknown"}
		}
		commentAuthors = append(commentAuthors, domain.CommentWithAuthor{
			Comment: *comment,
			Author: domain.PublicUser{
				ID:        commentAuthor.ID,
				Username:  commentAuthor.Username,
				CreatedAt: commentAuthor.CreatedAt.String(),
			},
		})
	}

	// Get current user
	user, _ := GetUserFromContext(r.Context())
	var publicUser *domain.PublicUser
	if user != nil {
		publicUser = &domain.PublicUser{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.String(),
		}
	}

	// Prepare data for template using your structs
	data := PostDetailViewData{
		BaseViewData: BaseViewData{
			CurrentUser: publicUser,
			Error:       "",
		},
		Post: domain.PostWithAuthor{
			Post:   *post,
			Author: domain.PublicUser{
				ID:        author.ID,
				Username:  author.Username,
				CreatedAt: author.CreatedAt.String(),
			},
		},
		Comments: commentAuthors,
		CommentForm: CommentForm{
			PostID: id,
			Body:   "",
		},
	}

	h.renderTemplate(w, "post_detail.html", data)
}

// HandleCreatePostPage displays the create post form
func (h *PostHandler) HandleCreatePostPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	categories, err := h.postRepo.GetAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		categories = []*domain.Category{}
	}

	publicUser := &domain.PublicUser{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.String(),
	}

	data := struct {
		BaseViewData
		Categories []*domain.Category
	}{
		BaseViewData: BaseViewData{
			CurrentUser: publicUser,
			Error:       "",
		},
		Categories: categories,
	}

	h.renderTemplate(w, "create_post.html", data)
}

// HandleCreatePost processes post creation
func (h *PostHandler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories"]

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Create post
	post := &domain.Post{
		Title:     title,
		Content:   content,
		UserID:    user.ID,
		Username:  user.Username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save post using repository
	if err := h.postRepo.CreatePost(post, categories); err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderTemplate renders an HTML template
func (h *PostHandler) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if err := h.templates.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf("Error rendering template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}