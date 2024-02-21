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
	TypeTranscodeLive  = "task:live"
	TypeTranscodeVideo = "task:video"
)

type TranscodeVideoPayload struct {
	Id  string
	Src string
}

type TranscodeLivePayload struct {
	Id      string
	App     string
	LiveKey string
}

func NewTranscodeVideoTask(id string, src string) (*asynq.Task, error) {
	payload, err := json.Marshal(TranscodeVideoPayload{Id: id, Src: src})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTranscodeVideo, payload), nil
}

func NewTranscodeLiveTask(id string, app string, liveKey string) (*asynq.Task, error) {
	payload, err := json.Marshal(TranscodeLivePayload{Id: id, App: app, LiveKey: liveKey})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTranscodeLive, payload), nil
}

func HandleTranscodeVideoTask(ctx context.Context, t *asynq.Task) error {
	var transVideo TranscodeVideoPayload
	if err := json.Unmarshal(t.Payload(), &transVideo); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	fileName := utils.GetFileName(transVideo.Src)
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

	id, err := db.NewLog("VIDEO", transVideo.Id, string(encode))
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("insert error: %v: %w", err, asynq.SkipRetry)
	}

	db.UpdateLogById(id, "PROCESS", "")
	err = utils.FFmpegIns.TranscodeVideo(fileName, transVideo.Src)
	if err != nil {
		db.UpdateLogById(id, "ERROR", err.Error())
		return fmt.Errorf("transcode error: %v: %w", err, asynq.SkipRetry)
	}
	db.UpdateLogById(id, "SUCCESS", "")

	err = db.UpdateSrcTranscode(transVideo.Id, fileName)
	if err != nil {
		return fmt.Errorf("transcode error: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf(" [*] transcode video %s SUCCESS", transVideo.Src)
	return nil
}

func HandleTranscodeLiveTask(ctx context.Context, t *asynq.Task) error {
	var transLive TranscodeLivePayload
	if err := json.Unmarshal(t.Payload(), &transLive); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	metadata := map[string]string{
		"videoType":   "flv",
		"encodeVideo": "H.264",
		"encodeAudio": "AAC",
		"resolution":  "360p, 480p, 720p, 1080p",
	}

	encode, err := json.Marshal(metadata)
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	id, err := db.NewLog("LIVE", transLive.Id, string(encode))
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("insert error: %v: %w", err, asynq.SkipRetry)
	}

	db.UpdateLogById(id, "PROCESS", "")
	err = utils.FFmpegIns.TranscodeLive(transLive.App, transLive.LiveKey)
	if err != nil {
		db.UpdateLogById(id, "ERROR", err.Error())
		return fmt.Errorf("transcode error: %v: %w", err, asynq.SkipRetry)
	}
	db.UpdateLogById(id, "SUCCESS", "")
	log.Printf(" [*] transocde live %s SUCCESS", transLive.LiveKey)
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
	mux.HandleFunc(TypeTranscodeLive, HandleTranscodeLiveTask)

	if err := srv.Run(mux); err != nil {
		fmt.Println(err)
	}
}
