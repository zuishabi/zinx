package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"strconv"
	"zinx/GodQQ/mysqlQQ"
)

type FFProbeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func InitVideoModule() {
	//获取所有的视频
	directory := "./videos/video_sources"
	files, err := os.Open(directory)
	if err != nil {
		fmt.Println("error opening directory:", err)
		return
	}
	defer files.Close()
	fileInfos, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return
	}
	for _, fileInfos := range fileInfos {
		fmt.Println(fileInfos.Name())
		videoInfo := mysqlQQ.VideoList{}
		err = mysqlQQ.Db.Where("video_name = ?", fileInfos.Name()).First(&videoInfo).Error
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			videoInfo.VideoLen = GetLength(directory + "/" + fileInfos.Name())
			videoInfo.VideoName = fileInfos.Name()
			mysqlQQ.Db.Create(&videoInfo)
			mysqlQQ.Db.Where("video_name = ?", fileInfos.Name()).First(&videoInfo)
			os.MkdirAll("./videos/"+strconv.Itoa(int(videoInfo.ID)), os.ModePerm)
			CutVideo(directory+"/"+fileInfos.Name(), "./videos/"+strconv.Itoa(int(videoInfo.ID)))
		}
	}
}

// 将视频进行分段
//
//	func CutVideo(url string, start int, output string) error {
//		ffmpegPath := "C:\\ffmpeg-7.1-essentials_build\\bin\\ffmpeg.exe"
//		// 执行
//		//cmd := exec.Command(ffmpegPath, "-i", url, "-ss", strconv.Itoa(start), "-t", "8", "-c", "copy", output)
//		//cmd := exec.Command(ffmpegPath, "-ss", strconv.Itoa(start), "-i", url, "-t", "8", "-c", "copy", "-copyts", output)
//		cmd := exec.Command(ffmpegPath, "-ss", strconv.Itoa(start), "-to", strconv.Itoa(start+4), "-accurate_seek", "-i", url, "-c:v", "libx264", "-c:a", "aac", "-avoid_negative_ts", "1", "-y", output)
//		// Capture standard output and standard error
//		var out bytes.Buffer
//		var stderr bytes.Buffer
//		cmd.Stdout = &out
//		cmd.Stderr = &stderr
//
//		// Run the command
//		err := cmd.Run()
//		if err != nil {
//			log.Fatalf("Error processing videos: %v: %v", err, stderr.String())
//			return err
//		}
//		return nil
//	}
func CutVideo(url string, outputDir string) error {
	ffmpegPath := "C:\\ffmpeg-7.1-essentials_build\\bin\\ffmpeg.exe"
	// 使用 ffmpeg 的 segment 模式进行切片
	cmd := exec.Command(ffmpegPath, "-i", url, "-c:v", "libx264", "-c:a", "aac", "-f", "segment", "-segment_time", "8", "-reset_timestamps", "1", outputDir+"/%03d.mp4")

	// Capture standard output and standard error
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error processing videos: %v: %v", err, stderr.String())
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
