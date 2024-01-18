package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"transcode/db"
	"transcode/task"
	"transcode/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func main() {
	loadConfig()
	db.ConnectPostgresql()
	app := fiber.New()
	task.ConnectRedis()
	utils.NewFFmpegIns()
	defer task.ClienTask.Close()
	go func() { task.StartWorker() }()

	app.Post("/api/v1/live", func(c *fiber.Ctx) error {
		// task.AddTaskTranscodeStream()
		payload := struct {
			LiveKey string `json:"liveKey"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}

		var id string
		var status string
		err := db.Connect.QueryRow(context.Background(), "SELECT id, status FROM livestreams WHERE \"liveKey\" = $1", payload.LiveKey).Scan(&id, &status)
		if err != nil {
			fmt.Print(err)
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}
		return c.JSON(fiber.Map{"status": 200, "error": nil, "data": fiber.Map{"id:": id, "liveKey": payload.LiveKey, "status": status}})
	})

	app.Post("/api/v1/video", func(c *fiber.Ctx) error {
		payload := struct {
			Id string `json:"id"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}

		var src string
		err := db.Connect.QueryRow(context.Background(), "SELECT src FROM videos WHERE id = $1", payload.Id).Scan(&src)
		if err != nil {
			fmt.Print(err)
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}

		task.AddTaskTranscodeVideo(payload.Id, src)
		return c.JSON(fiber.Map{"status": 200, "error": nil, "data": nil})
	})

	GracefullyShutdown(app)
	app.Listen(":3002")
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
