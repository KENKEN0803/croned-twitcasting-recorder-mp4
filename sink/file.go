package sink

import (
	"fmt"
	"github.com/jzhang046/croned-twitcasting-recorder/record"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	timeFormat        = "20060102-1504"
	sinkChanBuffer    = 16
	baseRecordingPath = "./file"
)

var IsTerminating = false

func sanitizePathString(input string) string {
	regex := regexp.MustCompile(`[\{\}\[\]\/?.,;:|\)*~!^\-_+<>@\#$%&\\\=\(\'\"\n\r]+`)
	specialRemoved := regex.ReplaceAllString(input, "")
	fullSpaceRemoved := strings.ReplaceAll(specialRemoved, "　", "_")
	halfSpaceRemoved := strings.ReplaceAll(fullSpaceRemoved, " ", "_")
	return halfSpaceRemoved
}

func GetFilePaths(recordCtx record.RecordContext) (string, string, string) {
	timestamp := time.Now().Format(timeFormat)
	streamer := sanitizePathString(recordCtx.GetStreamer())
	streamTitle := sanitizePathString(recordCtx.GetStreamTitle())

	fileName := fmt.Sprintf("%s-%s", timestamp, streamTitle)
	streamerRecordPath := fmt.Sprintf("%s/%s", baseRecordingPath, streamer)
	tsFilePath := fmt.Sprintf("%s/%s.ts", streamerRecordPath, fileName)
	mp4FilePath := fmt.Sprintf("%s/%s.mp4", streamerRecordPath, fileName)
	return tsFilePath, mp4FilePath, streamerRecordPath
}

func CreateRecordingFolder(streamerRecordPath string) error {
	if _, err := os.Stat(baseRecordingPath); os.IsNotExist(err) {
		err = os.Mkdir(baseRecordingPath, 0755)
		if err != nil {
			log.Printf("Error creating recording folder %s: %v\n", baseRecordingPath, err)
			return err
		}
	}

	if _, err := os.Stat(streamerRecordPath); os.IsNotExist(err) {
		err = os.Mkdir(streamerRecordPath, 0755)
		if err != nil {
			log.Printf("Error creating recording folder %s: %v\n", streamerRecordPath, err)
			return err
		}
	}
	return nil
}

func NewFileSink(recordCtx record.RecordContext) (chan<- []byte, error) {
	tsFilePath, mp4FilePath, streamerRecordPath := GetFilePaths(recordCtx)

	err := CreateRecordingFolder(streamerRecordPath)
	if err != nil {
		return nil, err
	}

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(tsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	log.Printf("Recording file %s", tsFilePath)

	sinkChan := make(chan []byte, sinkChanBuffer)

	go func() {
		defer f.Close()
		for data := range sinkChan {
			if _, err = f.Write(data); err != nil {
				log.Printf("Error writing recording file %s: %v\n", tsFilePath, err)
				recordCtx.Cancel()
				return
			}
		}

		log.Printf("Completed writing all data to %s\n", tsFilePath)

		if !IsTerminating {
			go func() {
				if isFFmpegInstalled() {
					err := convertTsToMp4(tsFilePath, mp4FilePath, recordCtx.GetEncodeOption())
					if err == nil {
						_ = RemoveFile(tsFilePath)
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
