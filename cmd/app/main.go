package main

import (
	"github.com/EmelinDanila/go_test/internal/cache"
	"github.com/EmelinDanila/go_test/internal/db"
	"github.com/EmelinDanila/go_test/internal/goods"
	"github.com/EmelinDanila/go_test/internal/logger"
	"github.com/EmelinDanila/go_test/internal/nats"
	"github.com/gofiber/fiber/v2"
)

func main() {
	db.InitPostgres()
	cache.InitRedis()
	nats.Init()

	go logger.StartConsumer()

	app := fiber.New()
	api := app.Group("/goods")
	goods.Routes(api)

	app.Listen(":8080")
}
