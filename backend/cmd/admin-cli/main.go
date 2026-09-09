package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/imageforge/imageforge/ent/user"
	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/pkg/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Load config
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create a simple Ent client wrapper
	db := &dbClient{pool: pool}

	switch command {
	case "promote":
		if len(os.Args) < 3 {
			fmt.Println("Usage: admin-cli promote <email>")
			os.Exit(1)
		}
		if err := promoteUser(ctx, db, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "demote":
		if len(os.Args) < 3 {
			fmt.Println("Usage: admin-cli demote <email>")
			os.Exit(1)
		}
		if err := demoteUser(ctx, db, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := listAdmins(ctx, db); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "gen-key":
		// Generate a new AES-256 key for IF_CREDENTIAL_KEY.
		key, err := crypto.GenerateKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(key)
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints CLI usage information.
func printUsage() {
	fmt.Println("ImageForge Admin CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  admin-cli promote <email>   Promote a user to admin")
	fmt.Println("  admin-cli demote <email>    Demote an admin to user")
	fmt.Println("  admin-cli list              List all admins")
	fmt.Println("  admin-cli gen-key           Generate AES-256 key for IF_CREDENTIAL_KEY")
}

// dbClient is a simple database client for the CLI tool.
type dbClient struct {
	pool *pgxpool.Pool
}

func (d *dbClient) queryOne(ctx context.Context, sql string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := d.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned")
	}

	values, err := rows.Values()
	if err != nil {
		return nil, err
	}

	fields := rows.FieldDescriptions()
	result := make(map[string]interface{})
	for i, f := range fields {
		result[f.Name] = values[i]
	}
	return result, nil
}

func (d *dbClient) queryAll(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := d.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		fields := rows.FieldDescriptions()
		result := make(map[string]interface{})
		for i, f := range fields {
			result[f.Name] = values[i]
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (d *dbClient) exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := d.pool.Exec(ctx, sql, args...)
	return err
}

func promoteUser(ctx context.Context, db *dbClient, email string) error {
	// Check if user exists
	result, err := db.queryOne(ctx, "SELECT id, email, role FROM users WHERE email = $1", email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if result["role"] == "admin" {
		fmt.Printf("User %s is already an admin\n", email)
		return nil
	}

	// Promote to admin
	err = db.exec(ctx, "UPDATE users SET role = $1, updated_at = $2 WHERE email = $3", "admin", time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to promote user: %w", err)
	}

	fmt.Printf("Successfully promoted %s to admin\n", email)
	return nil
}

func demoteUser(ctx context.Context, db *dbClient, email string) error {
	// Check if user exists
	result, err := db.queryOne(ctx, "SELECT id, email, role FROM users WHERE email = $1", email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if result["role"] == "user" {
		fmt.Printf("User %s is already a regular user\n", email)
		return nil
	}

	// Demote to user
	err = db.exec(ctx, "UPDATE users SET role = $1, updated_at = $2 WHERE email = $3", "user", time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to demote user: %w", err)
	}

	fmt.Printf("Successfully demoted %s to user\n", email)
	return nil
}

func listAdmins(ctx context.Context, db *dbClient) error {
	results, err := db.queryAll(ctx, "SELECT id, email, name, role, status, created_at FROM users WHERE role = $1 ORDER BY created_at", "admin")
	if err != nil {
		return fmt.Errorf("failed to list admins: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No admin users found")
		return nil
	}

	fmt.Println("Admin users:")
	fmt.Println("------------")
	for _, r := range results {
		fmt.Printf("  ID: %v, Email: %v, Name: %v, Status: %v, Created: %v\n",
			r["id"], r["email"], r["name"], r["status"], r["created_at"])
	}
	return nil
}

// Ensure user package is referenced (for role constants).
var _ = user.RoleAdmin
