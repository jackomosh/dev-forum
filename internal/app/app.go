package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"forum/internal/config"
	"forum/internal/handler"
	"forum/internal/repository"
	"forum/internal/repository/sqlite"
)

type Application struct {
	Config     config.Config
	Repository repository.Repository
	server     *http.Server
}

type Dependencies struct {
	Config config.Config
}

func Run() error {
	application, err := New(Dependencies{Config: config.Default()})
	if err != nil {
		return err
	}
	defer application.Repository.Close()

	log.Printf("server running at http://localhost:%s", application.Config.Server.Port)
	return application.ListenAndServe()
}

func New(deps Dependencies) (*Application, error) {
	cfg := deps.Config
	if cfg.Server.Port == "" {
		cfg = config.Default()
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "forum.db"
	}

	client, err := sqlite.NewClient(cfg.Database.Path)
	if err != nil {
		return nil, err
	}

	if err := applySchema(client); err != nil {
		_ = client.Close()
		return nil, err
	}

	repo := sqlite.NewRepository(client)
	renderer := handler.NewRenderer("web/templates")
	forumHandler := handler.NewForumHandler(repo, renderer, handler.Options{
		SessionCookieName: cfg.Session.CookieName,
		SessionDuration:   cfg.Session.Duration,
		SessionSecure:     cfg.Session.Secure,
		SessionSameSite:   cfg.Session.SameSite,
		PasswordMinLength: cfg.Security.PasswordMinLength,
	})

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	forumHandler.RegisterRoutes(mux)

	return &Application{
		Config:     cfg,
		Repository: repo,
		server: &http.Server{
			Addr:         serverAddress(cfg),
			Handler:      mux,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}, nil
}

func (a *Application) ListenAndServe() error {
	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func applySchema(client *sqlite.Client) error {
	schemaPath := "schema.sql"

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %q: %w", schemaPath, err)
	}

	if _, err := client.DB().Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	if err := seedData(client); err != nil {
		return fmt.Errorf("seed data: %w", err)
	}

	return nil
}

func serverAddress(cfg config.Config) string {
	if cfg.Server.Host == "" {
		return ":" + cfg.Server.Port
	}
	return net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
}

func seedData(client *sqlite.Client) error {
	ctx := context.Background()
	db := client.DB()

	var userCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		return err
	}
	if userCount > 0 {
		return nil
	}

	now := time.Now().Unix()

	users := []struct {
		username string
		email    string
		password string
		role     string
	}{
		{"alex_dev", "alex@example.com", "password123", "admin"},
		{"sarah_codes", "sarah@example.com", "password123", "member"},
		{"mike_arch", "mike@example.com", "password123", "member"},
		{"lisa_go", "lisa@example.com", "password123", "member"},
	}

	userIDs := make([]int64, 0, len(users))
	for _, u := range users {
		hash, err := hashPassword(u.password)
		if err != nil {
			return err
		}
		res, err := db.ExecContext(ctx,
			"INSERT INTO users (username, email, password_hash, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			u.username, u.email, hash, u.role, now, now,
		)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		userIDs = append(userIDs, id)
	}

	posts := []struct {
		title   string
		body    string
		author  int64
		cats    []int
		status  string
	}{
		{
			title: "Structuring Go Architectures for High-Concurrency Applications",
			body:  "Building performance-focused servers requires deep insight into thread schedulers, Go's runtime behavior, and memory allocation optimization. In this writeup, we analyze custom middleware layers, context handling, and high-performance routing layouts designed for microsecond-scale latencies. We explore how goroutine pools, worker patterns, and channel-based communication can transform a standard HTTP service into a high-throughput system capable of handling tens of thousands of concurrent connections with minimal allocation overhead.",
			author: userIDs[0],
			cats:   []int{1},
			status: "published",
		},
		{
			title: "An Introduction to Modern Web Assembly Runtimes",
			body:  "Unveiling performance differences when deploying system binaries onto modern cloud instances using container-less WASM sandboxes. We compare Wasmtime, Wasmer, and WasmEdge runtimes, measuring cold-start latency, memory footprint, and JIT compilation efficiency. The findings reveal that WASM runtimes can achieve near-native performance for compute-intensive workloads while maintaining strict security boundaries.",
			author: userIDs[1],
			cats:   []int{1},
			status: "published",
		},
		{
			title: "Containerization Checklist: Production Standards",
			body:  "Hardening Docker environments with rootless executions, structural image layers, multi-stage assemblies, and minimal alpine binaries. This guide covers security hardening techniques, image scanning best practices, and optimization strategies that reduce attack surface while keeping image sizes under 10MB for microservices deployments.",
			author: userIDs[2],
			cats:   []int{1},
			status: "published",
		},
		{
			title: "Designing Clean and Resilient REST APIs",
			body:  "How standardizing interfaces, utilizing appropriate state codes, and maintaining explicit contract types mitigates downstream bugs. We discuss HATEOAS principles, API versioning strategies, and the importance of consistent error response formats. Real-world examples demonstrate how proper API design reduces integration failures and improves developer experience.",
			author: userIDs[3],
			cats:   []int{1},
			status: "published",
		},
	}

	for _, p := range posts {
		res, err := db.ExecContext(ctx,
			"INSERT INTO posts (author_id, title, body, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			p.author, p.title, p.body, p.status, now, now,
		)
		if err != nil {
			return err
		}
		postID, _ := res.LastInsertId()
		for _, catID := range p.cats {
			if _, err := db.ExecContext(ctx, "INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID); err != nil {
				return err
			}
		}
	}

	return nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	key := pbkdf2SHA256([]byte(password), salt, 120000, 32)
	return fmt.Sprintf("pbkdf2_sha256$120000$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := sha256.Size
	blockCount := (keyLen + hashLen - 1) / hashLen
	derived := make([]byte, 0, blockCount*hashLen)

	for block := 1; block <= blockCount; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)

		var blockIndex [4]byte
		binary.BigEndian.PutUint32(blockIndex[:], uint32(block))
		mac.Write(blockIndex[:])

		u := mac.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)

		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)

			for j := range t {
				t[j] ^= u[j]
			}
		}

		derived = append(derived, t...)
	}

	return derived[:keyLen]
}
