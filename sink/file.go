package sink

import (
	"fmt"
	"github.com/jzhang046/croned-twitcasting-recorder/record"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	timeFormat     = "20060102-1504"
	sinkChanBuffer = 16
)

var IsTerminating = false

func GetFileNames(recordCtx record.RecordContext) (string, string, string) {
	timestamp := time.Now().Format(timeFormat)
	streamer := recordCtx.GetStreamer()

	// Replace ":" with "_"
	streamer = strings.ReplaceAll(streamer, ":", "_")

	baseFilename := fmt.Sprintf("%s-%s", streamer, timestamp)
	tsFilename := baseFilename + ".ts"
	mp4Filename := baseFilename + ".mp4"
	return baseFilename, tsFilename, mp4Filename
}

func NewFileSink(recordCtx record.RecordContext) (chan<- []byte, error) {
	// If the file doesn't exist, create it, or append to the file
	_, tsFilename, mp4Filename := GetFileNames(recordCtx)
	f, err := os.OpenFile(tsFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	log.Printf("Recording file %s", tsFilename)

	sinkChan := make(chan []byte, sinkChanBuffer)

	go func() {
		defer f.Close()
		for data := range sinkChan {
			if _, err = f.Write(data); err != nil {
				log.Printf("Error writing recording file %s: %v\n", tsFilename, err)
				recordCtx.Cancel()
				return
			}
		}

		log.Printf("Completed writing all data to %s\n", tsFilename)

		if !IsTerminating {
			go func() {
				if isFFmpegInstalled() {
					err := convertTsToMp4(tsFilename, mp4Filename, recordCtx.GetEncodeOption())
					if err == nil {
						_ = RemoveFile(tsFilename)
					}
				} else {
					log.Printf("ffmpeg is not installed, skipping conversion to mp4\n")
				}
			}()
		}

	}()

	return sinkChan, nil
}

func isFFmpegInstalled() bool {
	ffmpegCmd := exec.Command("ffmpeg", "-version")
	err := ffmpegCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && !exitErr.Success() {
			return false
		}
	}
	return true
}

func convertTsToMp4(tsFilename string, mp4Filename string, encodeOption *string) error {
	// Run ffmpeg command to convert .ts to .mp4

	if encodeOption == nil {
		encodeOption = new(string)
		*encodeOption = "copy"
	}

	// Split the encodeOption string into separate arguments
	encodeOptions := strings.Fields(*encodeOption)

	// Create a command with individual arguments
	ffmpegArgs := []string{"-i", tsFilename, "-c:v"}
	ffmpegArgs = append(ffmpegArgs, encodeOptions...)
	ffmpegArgs = append(ffmpegArgs, "-c:a", "copy", mp4Filename)

	log.Printf("Stert Converting... ffmpeg args = %v\n", ffmpegArgs)

	ffmpegCmd := exec.Command("ffmpeg", ffmpegArgs...)

	err := ffmpegCmd.Run()
	if err != nil {
		log.Printf("Error running ffmpeg command: %v\n", err)
		return err
	}
	log.Printf("Conversion to %s completed\n", mp4Filename)
	return nil
}

func RemoveFile(filename string) error {
	if err := os.Remove(filename); err != nil {
		log.Printf("Error removing %s: %v\n", filename, err)
		return err
	}
	log.Printf("Removed %s\n", filename)

	return nil
}
