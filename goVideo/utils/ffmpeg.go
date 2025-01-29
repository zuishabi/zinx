package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

type FFProbeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// 初始化视频，将完整的视频进行切片，并存储
func InitVideo(url string) {
	length := GetLength(url)
	fmt.Println(length)
	//计算剪切出的视频数量
	cutTime := int((length / 8) + 1)
	for i := 0; i < cutTime; i++ {
		CutVideo(url, i*8, url+"_"+strconv.Itoa(i)+".mp4")
	}
}

// 将视频进行分段
func CutVideo(url string, start int, output string) error {
	ffmpegPath := "C:\\ffmpeg-7.1-essentials_build\\bin\\ffmpeg.exe"
	// 执行
	//cmd := exec.Command(ffmpegPath, "-i", url, "-ss", strconv.Itoa(start), "-t", "8", "-c", "copy", output)
	//cmd := exec.Command(ffmpegPath, "-ss", strconv.Itoa(start), "-i", url, "-t", "8", "-c", "copy", "-copyts", output)
	cmd := exec.Command(ffmpegPath, "-ss", strconv.Itoa(start), "-to", strconv.Itoa(start+8), "-accurate_seek", "-i", url, "-c:v", "libx264", "-c:a", "aac", "-avoid_negative_ts", "1", "-y", output)
	// Capture standard output and standard error
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error processing video: %v: %v", err, stderr.String())
		return err
	}
	return nil
}

// 获取传入的视频的长度
func GetLength(url string) float64 {
	// Specify the full path to the ffprobe executable
	ffprobePath := "C:\\ffmpeg-7.1-essentials_build\\bin\\ffprobe.exe"

	// Build the ffprobe command
	cmd := exec.Command(ffprobePath, "-v", "error", "-show_format", "-show_streams", "-of", "json", url)

	// Capture standard output and standard error
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error running ffprobe: %v: %v", err, stderr.String())
	}

	// Parse the JSON output
	var ffprobeOutput FFProbeOutput
	err = json.Unmarshal(out.Bytes(), &ffprobeOutput)
	if err != nil {
		log.Fatalf("Error parsing ffprobe output: %v", err)
	}
	res, err := strconv.ParseFloat(ffprobeOutput.Format.Duration, 64)
	if err != nil {
		fmt.Println(err)
	}
	return res
}
