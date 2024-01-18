package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"transcode/db"
	"transcode/utils"

	"github.com/spf13/viper"

	"github.com/hibiken/asynq"
)

const (
	TypeTranscodeStream = "task:stream"
	TypeTranscodeVideo  = "task:video"
)

type TranscodeVideoPayload struct {
	Id  string
	Src string
}

func NewTranscodeVideoTask(id string, src string) (*asynq.Task, error) {
	payload, err := json.Marshal(TranscodeVideoPayload{Id: id, Src: src})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTranscodeVideo, payload), nil
}

func HandleTranscodeVideoTask(ctx context.Context, t *asynq.Task) error {
	var transVideo TranscodeVideoPayload
	if err := json.Unmarshal(t.Payload(), &transVideo); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	metadata := map[string]string{
		"videoType":   "hls",
		"encodeVideo": "H.264",
		"encodeAudio": "AAC",
		"resolution":  "360p, 480p, 720p, 1080p",
	}

	encode, err := json.Marshal(metadata)
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	err = db.NewLog("VIDEO", string(encode))
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("insert error: %v: %w", err, asynq.SkipRetry)
	}

	utils.FFmpegIns.TranscodeVideo(transVideo.Src)
	println(transVideo.Src)
	log.Printf(" [*] stream %s", transVideo.Src)
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
	mux.HandleFunc(TypeTranscodeVideo, HandleTranscodeVideoTask)

	if err := srv.Run(mux); err != nil {
		fmt.Println(err)
	}
}
