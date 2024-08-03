package main

import (
	"os"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
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
	app := fiber.New()

	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Get("/health-check", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
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