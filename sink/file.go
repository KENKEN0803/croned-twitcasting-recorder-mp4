package sink

import (
	"fmt"
	"github.com/jzhang046/croned-twitcasting-recorder/record"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	timeFormat     = "20060102-1504"
	sinkChanBuffer = 16
)

func NewFileSink(recordCtx record.RecordContext) (chan<- []byte, error) {
	// If the file doesn't exist, create it, or append to the file
	filename := fmt.Sprintf("%s-%s.ts", recordCtx.GetStreamer(), time.Now().Format(timeFormat))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	log.Printf("Recording file %s", filename)

	sinkChan := make(chan []byte, sinkChanBuffer)

	go func() {
		defer f.Close()
		for data := range sinkChan {
			if _, err = f.Write(data); err != nil {
				log.Printf("Error writing recording file %s: %v\n", filename, err)
				recordCtx.Cancel()
				return
			}
		}
		log.Printf("Completed writing all data to %s\n", filename)

		if !isFFmpegInstalled() {
			log.Printf("ffmpeg is not installed, skipping conversion to mp4\n")
			return
		}

		_ = convertAndRemoveFile(filename)
		_ = removeFile(filename)
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

func convertAndRemoveFile(filename string) error {
	// Run ffmpeg command to convert .ts to .mp4
	mp4Filename := fmt.Sprintf("%s.mp4", filename)
	log.Printf("Converting %s to %s\n", filename, mp4Filename)
	ffmpegCmd := exec.Command("ffmpeg", "-i", filename, "-c:v", "copy", "-c:a", "copy", mp4Filename)
	err := ffmpegCmd.Run()
	if err != nil {
		log.Printf("Error running ffmpeg command: %v\n", err)
		return err
	}
	log.Printf("Conversion to %s completed\n", mp4Filename)
	return nil
}

func removeFile(filename string) error {
	if err := os.Remove(filename); err != nil {
		log.Printf("Error removing %s: %v\n", filename, err)
		return err
	}
	log.Printf("Removed %s\n", filename)

	return nil
}
