package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ViranjPatel/takt-visualiser/internal/api"
	"github.com/ViranjPatel/takt-visualiser/internal/db"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Database connection
	dbConn, err := db.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbConn.Close()

	// Redis connection
	redisClient := db.NewRedisClient(os.Getenv("REDIS_URL"))
	defer redisClient.Close()

	// Router setup
	r := mux.NewRouter()
	api.SetupRoutes(r, dbConn, redisClient)

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      c.Handler(r),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Wait for interrupt
	c1 := make(chan os.Signal, 1)
	signal.Notify(c1, os.Interrupt)
	<-c1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server shutting down")
}
