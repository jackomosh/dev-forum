// /home/stathuita/Desktop/forum/internal/repository/sqlite/repository_test.go
package sqlite

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"forum/internal/domain"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// TestNewClient tests database client creation
func TestNewClient(t *testing.T) {
	// Test: Create in-memory client
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Error("Expected client to not be nil")
	}

	if client.db == nil {
		t.Error("Expected db connection to not be nil")
	}

	// Verify connection works
	err = client.db.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

// TestClient_Migrations tests that schema can be applied
func TestClient_Migrations(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// Run migrations
	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	// Verify tables were created
	tables := []string{"users", "sessions", "posts", "categories", "comments", "votes"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err = client.db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Fatalf("Query for table %s failed: %v", table, err)
		}
		if count == 0 {
			t.Errorf("Table %s was not created", table)
		}
	}
}

// TestUserRepository_Create tests user creation
func TestUserRepository_Create(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Test: Create user
	user := &domain.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set, got 0")
	}
}

// TestUserRepository_GetByEmail tests retrieving a user by email
func TestUserRepository_GetByEmail(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Create a user first
	user := &domain.User{
		Username:     "getuser",
		Email:        "getuser@example.com",
		PasswordHash: "hashedpassword",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test: Get by email
	found, err := repo.GetByEmail(context.Background(), "getuser@example.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	if found == nil {
		t.Error("Expected to find user, got nil")
	}

	if found.Email != "getuser@example.com" {
		t.Errorf("Expected email 'getuser@example.com', got '%s'", found.Email)
	}
}

// TestUserRepository_GetByUsername tests retrieving a user by username
func TestUserRepository_GetByUsername(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Create a user first
	user := &domain.User{
		Username:     "uniqueuser",
		Email:        "unique@example.com",
		PasswordHash: "hashedpassword",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test: Get by username
	found, err := repo.GetByUsername(context.Background(), "uniqueuser")
	if err != nil {
		t.Fatalf("GetByUsername failed: %v", err)
	}

	if found == nil {
		t.Error("Expected to find user, got nil")
	}

	if found.Username != "uniqueuser" {
		t.Errorf("Expected username 'uniqueuser', got '%s'", found.Username)
	}
}

// TestUserRepository_ExistsByEmail tests email existence check
func TestUserRepository_ExistsByEmail(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Create a user
	user := &domain.User{
		Username:     "existsuser",
		Email:        "exists@example.com",
		PasswordHash: "hashedpassword",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test: Exists by email - existing
	exists, err := repo.ExistsByEmail(context.Background(), "exists@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}

	if !exists {
		t.Error("Expected email to exist, got false")
	}

	// Test: Exists by email - non-existent
	exists, err = repo.ExistsByEmail(context.Background(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("ExistsByEmail failed: %v", err)
	}

	if exists {
		t.Error("Expected email to not exist, got true")
	}
}

// TestUserRepository_ExistsByUsername tests username existence check
func TestUserRepository_ExistsByUsername(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Create a user
	user := &domain.User{
		Username:     "uniquename",
		Email:        "uniquename@example.com",
		PasswordHash: "hashedpassword",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test: Exists by username - existing
	exists, err := repo.ExistsByUsername(context.Background(), "uniquename")
	if err != nil {
		t.Fatalf("ExistsByUsername failed: %v", err)
	}

	if !exists {
		t.Error("Expected username to exist, got false")
	}

	// Test: Exists by username - non-existent
	exists, err = repo.ExistsByUsername(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("ExistsByUsername failed: %v", err)
	}

	if exists {
		t.Error("Expected username to not exist, got true")
	}
}

// TestUserRepository_Update tests updating a user
func TestUserRepository_Update(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Create a user
	user := &domain.User{
		Username:     "updateuser",
		Email:        "update@example.com",
		PasswordHash: "oldhash",
		Role:         domain.UserRoleMember,
	}

	err = repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update user
	user.Username = "updateduser"
	user.Email = "updated@example.com"
	user.PasswordHash = "newhash"

	err = repo.Update(context.Background(), user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	found, err := repo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Username != "updateduser" {
		t.Errorf("Expected username 'updateduser', got '%s'", found.Username)
	}

	if found.Email != "updated@example.com" {
		t.Errorf("Expected email 'updated@example.com', got '%s'", found.Email)
	}
}

// TestUserRepository_GetByID_NotFound tests getting a non-existent user
func TestUserRepository_GetByID_NotFound(t *testing.T) {
	// Setup
	client, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	err = runMigrations(client.db)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	repo := NewUserRepository(client)

	// Test: Get non-existent user
	found, err := repo.GetByID(context.Background(), 99999)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found != nil {
		t.Error("Expected nil for non-existent user, got a user")
	}
}

// Helper function to run migrations
func runMigrations(db *sql.DB) error {
	// Read schema file from the project root
	schema, err := os.ReadFile("../../../schema.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}
