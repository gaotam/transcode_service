package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/hibiken/asynq"
)

const (
	TypeTranscodeStream = "task:stream"
)

type TranscodeStreamPayload struct {
	ServerRTMP string
	Channel    string
	StreamKey  string
}

func NewTranscodeStreamTask(serverRTMP string, channel string, streamKey string) (*asynq.Task, error) {
	payload, err := json.Marshal(TranscodeStreamPayload{ServerRTMP: serverRTMP, Channel: channel, StreamKey: streamKey})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTranscodeStream, payload), nil
}

func HandleTranscodeStreamTask(ctx context.Context, t *asynq.Task) error {
	var transStream TranscodeStreamPayload
	if err := json.Unmarshal(t.Payload(), &transStream); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf(" [*] stream %s - %s - %s", transStream.ServerRTMP, transStream.Channel, transStream.StreamKey)
	return nil
}

func StartWorker() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: viper.GetString("redis.host"), Password: viper.GetString("redis.password")},
		asynq.Config{
			Concurrency: viper.GetInt("task.concurrency"),
			Queues: map[string]int{
				"prioritize": 6,
				"default":    3,
				"low":        1,
			}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeTranscodeStream, HandleTranscodeStreamTask)

	if err := srv.Run(mux); err != nil {
		fmt.Println(err)
	}
}
