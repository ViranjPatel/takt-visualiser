package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// Connect establishes database connection with optimized settings
func Connect(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Connection pool settings for high performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewRedisClient creates Redis connection
func NewRedisClient(redisURL string) *redis.Client {
	opt, _ := redis.ParseURL(redisURL)
	return redis.NewClient(opt)
}
