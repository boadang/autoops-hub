package web

import (
	"autoops-hub/internal/config"
	"autoops-hub/internal/engine"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func StartServer(port string, configPath string) {
	app := fiber.New()

	// Cho phép React app gọi API qua CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// API Kích hoạt chuỗi tác vụ Worker Pool
	app.Post("/api/trigger", func(c *fiber.Ctx) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
		}

		var tasks []engine.Task
		for _, t := range cfg.Workflows {
			tasks = append(tasks, engine.Task{
				ID:      t.ID,
				Name:    t.Name,
				Command: t.Command,
				Args:    t.Args,
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		results := engine.RunWorkerPool(ctx, tasks, 2)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Thực thi workflows thành công từ React Dashboard!",
			"results": results,
		})
	})

	app.Listen(port)
}
