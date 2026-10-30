package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
	"spot-solution/internal/config"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

type DB struct {
	*sql.DB
}

func NewConnection(cfg *config.Config) (*DB, error) {
	dsn := cfg.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("database configuration is missing or incomplete (DATABASE_URL or DB_HOST/DB_USER/DB_NAME required)")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return &DB{db}, nil
}

func (db *DB) RunMigrations() error {
	migrationSQL, err := MigrationsFS.ReadFile("migrations/000001_init_schema.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read embedded migration file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration script: %w", err)
	}

	log.Println("Database schema migrations applied successfully")
	return nil
}

func (db *DB) SeedDefaultAdmin(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", cfg.DefaultAdminEmail).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking existing admin: %w", err)
	}

	if !exists {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("error hashing default admin password: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO users (email, password_hash, full_name, role, is_active)
			VALUES ($1, $2, $3, $4, $5)
		`, cfg.DefaultAdminEmail, string(hashedPassword), "System Administrator", "admin", true)

		if err != nil {
			return fmt.Errorf("error creating default admin user: %w", err)
		}
		log.Printf("Seeded default admin user: %s", cfg.DefaultAdminEmail)
	}

	return nil
}

func (db *DB) ExecTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
