package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

const (
    batchSize = 10000 // Adjust this based on your memory constraints
    totalRecords = 1000000
)

func main() {
    _ = godotenv.Load()

    app := fiber.New(fiber.Config{
        BodyLimit: 1 * 1024 * 1024, // Limit request size to 1MB
    })

    app.Get("/hello", func(c *fiber.Ctx) error {
        return c.SendString("Hello World")
    })

    app.Post("/insert-million", insertRecords)

    port := os.Getenv("PORT")
    if port == "" {
        port = "4001"
    }
    app.Listen(":" + port)
}

func insertRecords(c *fiber.Ctx) error {
    startTime := time.Now()

    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        return c.Status(500).SendString("DATABASE_URL not set in environment")
    }

    ctx := context.Background()
    pool, err := pgxpool.Connect(ctx, dbURL)
    if err != nil {
        return c.Status(500).SendString(fmt.Sprintf("Unable to connect to database: %v", err))
    }
    defer pool.Close()

    initialCount, err := getUserCount(ctx, pool)
    if err != nil {
        return c.Status(500).SendString(fmt.Sprintf("Error getting initial count: %v", err))
    }

    for i := 0; i < totalRecords; i += batchSize {
        batch := &pgx.Batch{}
        for j := 0; j < batchSize && i+j < totalRecords; j++ {
            id := uuid.New()
            email := fmt.Sprintf("user%d@test.com", i+j+1)
            batch.Queue("INSERT INTO users(id, email) VALUES($1, $2)", id, email)
        }

        br := pool.SendBatch(ctx, batch)
        err = br.Close()
        if err != nil {
            return c.Status(500).SendString(fmt.Sprintf("Error executing batch: %v", err))
        }

        // Optional: add a small delay to prevent overwhelming the database
        // time.Sleep(10 * time.Millisecond)
    }

    finalCount, err := getUserCount(ctx, pool)
    if err != nil {
        return c.Status(500).SendString(fmt.Sprintf("Error getting final count: %v", err))
    }

    duration := time.Since(startTime)
    return c.SendString(fmt.Sprintf("Insertion complete. Initial count: %d, Final count: %d, Records inserted: %d, Time taken: %v",
        initialCount, finalCount, finalCount-initialCount, duration))
}

func getUserCount(ctx context.Context, pool *pgxpool.Pool) (int, error) {
    var count int
    err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    return count, err
}