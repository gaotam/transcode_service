package task

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

var ClienTask *asynq.Client

func ConnectRedis() {
	ClienTask = asynq.NewClient(asynq.RedisClientOpt{Addr: viper.GetString("redis.host"), Password: viper.GetString("redis.password")})
}

func AddTaskTranscodeVideo(id string, src string) {
	task, err := NewTranscodeVideoTask(id, src)
	if err != nil {
		fmt.Println("could not create video task ", err)
	}
	_, err = ClienTask.Enqueue(task, asynq.Queue("prioritize"))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	log.Printf(" [*] Successfully enqueued task")
}

func AddTaskTranscodeLive(id string, app string, liveKey string) {
	task, err := NewTranscodeLiveTask(id, app, liveKey)
	if err != nil {
		fmt.Println("could not create live task ", err)
	}
	_, err = ClienTask.Enqueue(task, asynq.Queue("prioritize"))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}

	log.Printf(" [*] Successfully enqueued task")
}
