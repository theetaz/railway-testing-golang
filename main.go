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

func main() {
    err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
    }

    app := fiber.New()

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

    batch := &pgx.Batch{}
    for i := 0; i < 1000000; i++ {
        id := uuid.New()
        email := fmt.Sprintf("user%d@test.com", i+1)
        batch.Queue("INSERT INTO users(id, email) VALUES($1, $2)", id, email)
    }

    br := pool.SendBatch(ctx, batch)
    err = br.Close()
    if err != nil {
        return c.Status(500).SendString(fmt.Sprintf("Error executing batch: %v", err))
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