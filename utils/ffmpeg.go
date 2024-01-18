package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var FFmpegIns FFmpeg

type FFmpeg struct {
	PathUpload string
}

func NewFFmpegIns() {
	FFmpegIns = FFmpeg{
		PathUpload: viper.GetString("path_upload"),
	}
}

func (f FFmpeg) TranscodeVideo(src string) (err error) {
	uid := uuid.New().String()
	pathVideo := fmt.Sprintf("%s/%s", f.PathUpload, src)
	pathTranscode := fmt.Sprintf("%s/%s/", f.PathUpload, "transcodes")
	args := []string{
		"-i", pathVideo, "-map", "0:v:0", "-map", "0:a:0", "-map", "0:v:0", "-map", "0:a:0", "-map", "0:v:0", "-map", "0:a:0", "-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "libx264", "-c:a", "aac", "-g", "300", "-ar", "48000",
		"-filter:v:0", "scale=w=480:h=360", "-b:v:0", "800k", "-minrate", "400k", "-maxrate", "1000k", "-b:a:0", "64k", "-crf", "36",
		"-filter:v:1", "scale=w=640:h=480", "-b:v:1", "1500k", "-minrate", "500k", "-maxrate", "2000k", "-b:a:1", "128k", "-crf", "25",
		"-filter:v:2", "scale=w=1280:h=720", "-b:v:2", "3000k", "-minrate", "1500k", "-maxrate", "3000k", "-b:a:2", "128k", "-crf", "22",
		"-filter:v:3", "scale=w=1920:h=1080", "-b:v:3", "5000k", "-minrate", "3000k", "-maxrate", "6000k", "-b:a:3", "192k", "-crf", "20",
		"-var_stream_map", fmt.Sprintf("v:0,a:0,name:%s_360p v:1,a:1,name:%s_480p v:2,a:2,name:%s_720p v:3,a:3,name:%s_1080p", uid, uid, uid, uid),
		"-preset", "slow", "-hls_list_size", "0", "-threads", "0", "-f", "hls",
		"-hls_playlist_type", "event", "-hls_time", "3",
		"-hls_flags", "independent_segments",
		"-master_pl_name", fmt.Sprintf("%s_master.m3u8", uid), "-y", pathTranscode + "%v.m3u8",
	}

	cmd := exec.Command("ffmpeg", args...)
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stderr)
	var m string
	for scanner.Scan() {
		m = scanner.Text()
		fmt.Println(m)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return errors.New(m)
	} else {
		fmt.Println("Success")
	}
	return nil
}
