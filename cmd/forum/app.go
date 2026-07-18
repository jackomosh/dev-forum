package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"forum/internal/config"
	"forum/internal/handler"
	"forum/internal/repository/sqlite"
)

// Run starts the application
func Run() error {
	log.Println("=== Starting Forum Application ===")

	// Load config
	cfg := config.Default()
	log.Printf("Config loaded: Port=%s, DBPath=%s", cfg.Server.Port, cfg.Database.Path)

	// Initialize database client
	client, err := sqlite.NewClient(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer client.Close()
	log.Println("Database connected successfully")

	// Run migrations
	if err := applySchema(client); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Migrations completed successfully")

	// Create repository
	repo := sqlite.NewRepository(client)

	// Get individual repositories from the combined repository
	userRepo := repo.Users()
	sessionRepo := repo.Sessions()
	postRepo := repo.Posts()
	commentRepo := repo.Comments()
	voteRepo := repo.Votes()

	// Initialize auth service
	authService := handler.NewAuthService(userRepo, sessionRepo)
	log.Println("Auth service initialized")

	// Initialize renderer
	renderer := handler.NewRenderer("web/templates")
	log.Println("Renderer initialized")

	// Initialize middleware
	middleware := handler.NewMiddleware(authService)
	log.Println("Middleware initialized")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, renderer)
	postHandler := handler.NewPostHandler(postRepo, commentRepo, voteRepo, renderer)
	commentHandler := handler.NewCommentHandler(commentRepo, renderer)
	voteHandler := handler.NewVoteHandler(voteRepo, renderer)
	filterHandler := handler.NewFilterHandler(postRepo, renderer)
	log.Println("Handlers initialized")

	// Set up routes
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Public routes
	mux.HandleFunc("/", postHandler.HandleHomePage)
	mux.HandleFunc("/login", authHandler.HandleLoginPage)
	mux.HandleFunc("/login/submit", authHandler.HandleLogin)
	mux.HandleFunc("/register", authHandler.HandleRegisterPage)
	mux.HandleFunc("/register/submit", authHandler.HandleRegister)
	mux.HandleFunc("/logout", authHandler.HandleLogout)
	mux.HandleFunc("/post/", postHandler.HandleViewPost)

	// Filter routes
	mux.HandleFunc("/filter", filterHandler.HandleFilter)
	mux.HandleFunc("/api/categories", filterHandler.HandleGetCategories)

	// Protected routes (require authentication)
	mux.HandleFunc("/dashboard", middleware.RequireAuth(postHandler.HandleDashboard))
	mux.HandleFunc("/post/create", middleware.RequireAuth(postHandler.HandleCreatePostPage))
	mux.HandleFunc("/post/create/submit", middleware.RequireAuth(postHandler.HandleCreatePost))
	mux.HandleFunc("/comment/create", middleware.RequireAuth(commentHandler.HandleCreateComment))
	mux.HandleFunc("/vote", middleware.RequireAuth(voteHandler.HandleVote))

	// Vote status (public)
	mux.HandleFunc("/vote/status", voteHandler.HandleGetVoteStatus)

	log.Println("Routes set up successfully")

	// Create server
	serverAddr := ":" + cfg.Server.Port
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("✅ Server running at http://localhost%s", serverAddr)
	log.Println("Press Ctrl+C to stop")

	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Server shut down gracefully")
		return nil
	}
	return err
}

// applySchema runs database migrations
func applySchema(client *sqlite.Client) error {
	log.Println("Creating database schema...")
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'member',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT NOT NULL,
		status TEXT DEFAULT 'published',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS post_categories (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		body TEXT NOT NULL,
		status TEXT DEFAULT 'visible',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS votes (
		user_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		value INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		PRIMARY KEY (user_id, target_type, target_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);
	CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
	CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
	CREATE INDEX IF NOT EXISTS idx_comments_author_id ON comments(author_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	CREATE INDEX IF NOT EXISTS idx_votes_target ON votes(target_type, target_id);
	`

	_, err := client.DB().Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database schema created successfully")
	return nil
}
