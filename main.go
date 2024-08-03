package main

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type HealthCheck struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
	Port        string    `json:"port"`
	CPU         CPU       `json:"cpu"`
	Memory      Memory    `json:"memory"`
	Server      Server    `json:"server"`
}

type CPU struct {
	Usage     float64 `json:"usage"`
	Cores     int     `json:"cores"`
	ModelName string  `json:"modelName"`
}

type Memory struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	UsagePerc float64 `json:"usagePercentage"`
}

type Server struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	GoVer    string `json:"goVersion"`
	NumGo    int    `json:"numGoroutines"`
	Hostname string `json:"hostname"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
			fmt.Println("Error loading .env file")
	}
		
	app := fiber.New()

	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Get("/health-check", healthCheck)

	app.Post("/insert-million", insertMillionRecords)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4001"
	}
	app.Listen(":" + port)
}

func insertMillionRecords(c *fiber.Ctx) error {
	startTime := time.Now()

	// Database connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
			return c.Status(500).SendString("DATABASE_URL not set in environment")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error connecting to database: %v", err))
	}
	defer db.Close()

	// Create a prepared statement
	stmt, err := db.Prepare("INSERT INTO users(id, email) VALUES($1, $2)")
	if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error preparing statement: %v", err))
	}
	defer stmt.Close()

	// Number of goroutines to use
	numWorkers := 10
	recordsPerWorker := 10000 / numWorkers

	var wg sync.WaitGroup
	errChan := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
					defer wg.Done()

					for j := 0; j < recordsPerWorker; j++ {
							id := uuid.New()
							email := fmt.Sprintf("user%d@test.com", workerID*recordsPerWorker+j+1)
							_, err := stmt.Exec(id, email)
							if err != nil {
									errChan <- fmt.Errorf("error inserting record %d: %v", workerID*recordsPerWorker+j+1, err)
									return
							}
					}
			}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
			if err != nil {
					return c.Status(500).SendString(err.Error())
			}
	}

	duration := time.Since(startTime)
	return c.SendString(fmt.Sprintf("Successfully inserted 1 million records in %v", duration))
}

func healthCheck(c *fiber.Ctx) error {
	cpuUsage, err := cpu.Percent(time.Second, false)
	cpuUsageValue := 0.0
	if err == nil && len(cpuUsage) > 0 {
		cpuUsageValue = cpuUsage[0]
	}

	cpuInfo, err := cpu.Info()
	cpuModel := "Unknown"
	if err == nil && len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		vmStat = &mem.VirtualMemoryStat{}
	}

	hostname, _ := os.Hostname()

	health := HealthCheck{
		Status:      "OK",
		Timestamp:   time.Now(),
		Environment: os.Getenv("GO_ENV"),
		Port:        os.Getenv("PORT"),
		CPU: CPU{
			Usage:     cpuUsageValue,
			Cores:     runtime.NumCPU(),
			ModelName: cpuModel,
		},
		Memory: Memory{
			Total:     vmStat.Total,
			Used:      vmStat.Used,
			Available: vmStat.Available,
			UsagePerc: vmStat.UsedPercent,
		},
		Server: Server{
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			GoVer:    runtime.Version(),
			NumGo:    runtime.NumGoroutine(),
			Hostname: hostname,
		},
	}

	return c.JSON(health)
}