package task

import (
	"encoding/json"
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
		fmt.Println("could not create task ", err)
	}
	info, err := ClienTask.Enqueue(task, asynq.Queue("prioritize"))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}

	var transVideo TranscodeVideoPayload
	json.Unmarshal(info.Payload, &transVideo)
	log.Printf(" [*] Successfully enqueued task: %s", transVideo)
}
