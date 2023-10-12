package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	app := fiber.New()

	app.Get("/test", testHandler)
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

func testHandler(ctx *fiber.Ctx) error {
	agent := fiber.Get("http://notification-receiver.notifications-test/ready")
	s, r, e := agent.Bytes()
	fmt.Printf("---%v, ----%v, ----%e\n", s, r, e)
	return ctx.SendStatus(200)
}
