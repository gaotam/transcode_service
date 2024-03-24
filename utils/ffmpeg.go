package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var FFmpegIns FFmpeg

var resolutionMap = map[int]int{
	360:  1,
	480:  2,
	720:  3,
	1080: 4,
}

type FFmpeg struct {
	PathUpload string
}

func NewFFmpegIns() {
	FFmpegIns = FFmpeg{
		PathUpload: viper.GetString("path_upload"),
	}
}

func (f FFmpeg) GetResolution(src string) (result int, err error) {
	pathVideo := fmt.Sprintf("%s/%s", f.PathUpload, src)
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", pathVideo)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()

	parts := strings.Split(out.String(), "x")
	width, _ := strconv.Atoi(parts[0])
	height, _ := strconv.Atoi(strings.TrimSpace(parts[1]))

	if width >= 640 && width < 854 && height == 360 {
		return 360, nil
	}

	if width >= 854 && width < 1280 && height == 480 {
		return 480, nil
	}

	if width >= 1280 && width < 1920 && height == 720 {
		return 720, nil
	}

	if width >= 1920 && width < 2560 && height == 1080 {
		return 1080, nil
	}

	return 0, err
}

func (f FFmpeg) TranscodeVideo(id string, src string, resolution int) (err error) {
	pathVideo := fmt.Sprintf("%s/%s", f.PathUpload, src)
	pathTranscode := fmt.Sprintf("%s/%s/%s", f.PathUpload, "transcodes", id)

	err = CreateFolder(pathTranscode)
	if err != nil {
		return err
	}

	var args []string
	var_stream_map := ""

	args = append(args, "-i", pathVideo)

	for i := 0; i < resolutionMap[resolution]; i++ {
		args = append(args, "-map", "0:v:0", "-map", "0:a:0")
	}

	args = append(args, "-c:v", "libx264", "-c:a", "aac", "-g", "300", "-ar", "48000")

	if resolution >= 360 {
		args = append(args, "-filter:v:0", "scale=w=640:h=360", "-b:v:0", "800k", "-minrate", "400k", "-maxrate", "1000k", "-b:a:0", "64k", "-crf", "36")
		var_stream_map += fmt.Sprintf("v:0,a:0,name:%s_360p", id)
	}

	if resolution >= 480 {
		args = append(args, "-filter:v:1", "scale=w=854:h=480", "-b:v:1", "1500k", "-minrate", "500k", "-maxrate", "2000k", "-b:a:1", "128k", "-crf", "25")
		var_stream_map += fmt.Sprintf(" v:1,a:1,name:%s_480p", id)
	}

	if resolution >= 720 {
		args = append(args, "-filter:v:2", "scale=w=1280:h=720", "-b:v:2", "3000k", "-minrate", "1500k", "-maxrate", "3000k", "-b:a:2", "128k", "-crf", "22")
		var_stream_map += fmt.Sprintf(" v:2,a:2,name:%s_720p", id)
	}

	if resolution >= 1080 {
		args = append(args, "-filter:v:3", "scale=w=1920:h=1080", "-b:v:3", "5000k", "-minrate", "3000k", "-maxrate", "6000k", "-b:a:3", "192k", "-crf", "20")
		var_stream_map += fmt.Sprintf(" v:3,a:3,name:%s_1080p", id)
	}

	args = append(args,
		"-var_stream_map", var_stream_map,
		"-preset", "slow", "-hls_list_size", "0", "-threads", "0", "-f", "hls",
		"-hls_playlist_type", "event", "-hls_time", "3",
		"-hls_flags", "independent_segments",
		"-master_pl_name", fmt.Sprintf("%s_master.m3u8", id), "-y", pathTranscode+"/%v.m3u8",
	)

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
		return nil
	}
}

func (f FFmpeg) TranscodeLive(app string, liveKey string) (err error) {
	rtmpSource := fmt.Sprintf("rtmp://%s:%d/%s/%s?secret=%s", viper.GetString("rtmp.host"), viper.GetInt("rtmp.port"), app, liveKey, viper.GetString("rtmp.secret"))
	rtmpTranscode := fmt.Sprintf("rtmp://%s:%d/t_live/%s", viper.GetString("rtmp.host"), viper.GetInt("rtmp.port"), liveKey)

	args := []string{
		"-i", rtmpSource, "-filter_complex", "[0:v]scale=w=480:h=360[v360];[0:v]scale=w=640:h=480[v480]; [0:v]scale=w=1280:h=720[v720];[0:v]scale=w=1920:h=1080[v1080]",
		"-g", "300", "-ar", "48000",
		"-map", "[v360]", "-map", "0:a:0", "-c:v:0", "libx264", "-c:a:0", "aac", "-b:v:0", "800k", "-maxrate", "1000k", "-b:a:0", "64k", "-crf", "36", "-preset", "slow", "-threads", "0", "-f", "flv", GenerateRtmpTranscode(rtmpTranscode, 360),
		"-map", "[v480]", "-map", "0:a:0", "-c:v:0", "libx264", "-c:a:0", "aac", "-b:v:0", "1500k", "-maxrate", "2000k", "-b:a:0", "128k", "-crf", "25", "-preset", "slow", "-threads", "0", "-f", "flv", GenerateRtmpTranscode(rtmpTranscode, 480),
		"-map", "[v720]", "-map", "0:a:0", "-c:v:0", "libx264", "-c:a:0", "aac", "-b:v:0", "3000k", "-maxrate", "4000k", "-b:a:0", "128k", "-crf", "22", "-preset", "slow", "-threads", "0", "-f", "flv", GenerateRtmpTranscode(rtmpTranscode, 720),
		"-map", "[v1080]", "-map", "0:a:0", "-c:v:0", "libx264", "-c:a:0", "aac", "-b:v:0", "5000k", "-maxrate", "6000k", "-b:a:0", "192k", "-crf", "20", "-preset", "slow", "-threads", "0", "-f", "flv", GenerateRtmpTranscode(rtmpTranscode, 1080),
	}

	cmd := exec.Command("ffmpeg", args...)
	fullCommand := strings.Join(cmd.Args, " ")
	fmt.Println("Full command:", fullCommand)
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
		return nil
	}
}

func GenerateRtmpTranscode(rtmpUrl string, resolution int) string {
	return fmt.Sprintf("%s_%dp", rtmpUrl, resolution)
}
