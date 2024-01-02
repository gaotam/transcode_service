package main

import (
	"fmt"
	"os"
	"os/signal"
	"transcode/task"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func main() {
	loadConfig()
	app := fiber.New()
	task.ConnectRedis()
	defer task.ClienTask.Close()
	go func() { task.StartWorker() }()

	app.Get("/add-task", func(c *fiber.Ctx) error {
		task.AddTaskTranscodeStream()
		return c.SendString("Hello, World!")
	})

	GracefullyShutdown(app)
	app.Listen(":3000")
}

func loadConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func GracefullyShutdown(app *fiber.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()
}
