package main

import (
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

	app.Post("/api/v1/lives", func(c *fiber.Ctx) error {
		payload := struct {
			LiveKey string `json:"liveKey"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}

		result, err := db.GetLiveByLiveKey(payload.LiveKey)
		if err != nil {
			fmt.Print(err)
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}

		task.AddTaskTranscodeLive(result.Id, result.App, payload.LiveKey)
		return c.JSON(fiber.Map{"status": 200, "error": nil, "data": fiber.Map{"id:": result.Id, "liveKey": payload.LiveKey, "status": result.Status}})
	})

	app.Post("/api/v1/videos", func(c *fiber.Ctx) error {
		payload := struct {
			Id string `json:"id"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.JSON(fiber.Map{"status": 400, "error": err.Error(), "data": nil})
		}

		var src string
		src, err := db.GetSrcVideoById(payload.Id)
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
