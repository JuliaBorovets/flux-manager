package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	app := fiber.New()

	app.Get("/healthy", healthyHandler)
	app.Get("/ready", readyHandler)

	log.Fatal(app.Listen(":8888"))
}

func healthyHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func readyHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}
