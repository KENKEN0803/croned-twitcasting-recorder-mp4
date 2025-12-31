package sink

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jzhang046/croned-twitcasting-recorder-mp4/record"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/uploader"
)

const (
	timeFormat        = "20060102-1504"
	sinkChanBuffer    = 16
	baseRecordingPath = "./file"
)

var IsTerminating = false

type FileSink struct {
	tsFilePath  string
	mp4FilePath string
	uploader    uploader.Uploader
	recordCtx   record.RecordContext
}

func sanitizePathString(input string) string {
	regex := regexp.MustCompile(`[\{\}\[\]\/?.,;:|\)*~!^\-_+<>@\#$%&\\\=\(\'\"\n\r]+`)
	specialRemoved := regex.ReplaceAllString(input, "")
	fullSpaceRemoved := strings.ReplaceAll(specialRemoved, "　", "_")
	halfSpaceRemoved := strings.ReplaceAll(fullSpaceRemoved, " ", "_")
	return halfSpaceRemoved
}

func chkMaxFilenameLength(input string) string {
	if len(input) > 250 {
		return input[:250]
	}
	return input
}

func GetFilePaths(recordCtx record.RecordContext) (string, string, string) {
	timestamp := time.Now().Format(timeFormat)
	streamer := sanitizePathString(recordCtx.GetStreamer())
	streamTitle := sanitizePathString(recordCtx.GetStreamTitle())

	fileName := chkMaxFilenameLength(fmt.Sprintf("%s-%s", timestamp, streamTitle))
	if recordCtx.IsMembershipStream() {
		fileName = "_" + fileName
	}

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

func NewFileSink(recordCtx record.RecordContext, uploader uploader.Uploader) (chan<- []byte, error) {
	tsFilePath, mp4FilePath, streamerRecordPath := GetFilePaths(recordCtx)

	err := CreateRecordingFolder(streamerRecordPath)
	if err != nil {
		return nil, err
	}

	sink := &FileSink{
		tsFilePath:  tsFilePath,
		mp4FilePath: mp4FilePath,
		uploader:    uploader,
		recordCtx:   recordCtx,
	}

	return sink.start(), nil
}

func (f *FileSink) start() chan<- []byte {
	// If the file doesn't exist, create it, or append to the file
	file, err := os.OpenFile(f.tsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Printf("Failed to open file %s: %v", f.tsFilePath, err)
		f.recordCtx.Cancel()
		return nil
	}
	log.Printf("Recording file %s", f.tsFilePath)

	sinkChan := make(chan []byte, sinkChanBuffer)

	go func() {
		defer file.Close()
		for data := range sinkChan {
			if _, err = file.Write(data); err != nil {
				log.Printf("Error writing recording file %s: %v\n", f.tsFilePath, err)
				f.recordCtx.Cancel()
				return
			}
		}

		log.Printf("Completed writing all data to %s", f.tsFilePath)
		f.uploadTS()

		if !IsTerminating {
			go f.convertAndUploadMP4()
		}

	}()

	return sinkChan
}

func (f *FileSink) uploadTS() {
	if f.uploader != nil {
		go func() {
			remotePath := "/ts/" + filepath.Base(f.tsFilePath)
			if err := f.uploader.Upload(f.tsFilePath, remotePath); err != nil {
				log.Printf("TS upload failed for %s: %v", f.tsFilePath, err)
			}
		}()
	}
}

func (f *FileSink) convertAndUploadMP4() {
	if isFFmpegInstalled() {
		err := f.convertTsToMp4()
		if err != nil {
			return // Conversion failed, so don't upload or remove
		}

		if f.uploader != nil {
			go func() {
				remotePath := "/mp4/" + filepath.Base(f.mp4FilePath)
				if err := f.uploader.Upload(f.mp4FilePath, remotePath); err != nil {
					log.Printf("MP4 upload failed for %s: %v", f.mp4FilePath, err)
				}
			}()
		}
		_ = RemoveFile(f.tsFilePath)
	} else {
		log.Printf("ffmpeg is not installed, skipping conversion to mp4\n")
	}
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

func (f *FileSink) convertTsToMp4() error {
	encodeOption := f.recordCtx.GetEncodeOption()
	if encodeOption == nil {
		defaultOption := "copy"
		encodeOption = &defaultOption
	}

	encodeOptions := strings.Fields(*encodeOption)

	ffmpegArgs := []string{"-i", f.tsFilePath, "-c:v"}
	ffmpegArgs = append(ffmpegArgs, encodeOptions...)
	ffmpegArgs = append(ffmpegArgs, "-c:a", "copy", f.mp4FilePath)

	log.Printf("Start Converting... ffmpeg args = %v", ffmpegArgs)

	ffmpegCmd := exec.Command("ffmpeg", ffmpegArgs...)

	err := ffmpegCmd.Run()
	if err != nil {
		log.Printf("Error running ffmpeg command: %v", err)
		return err
	}
	log.Printf("Conversion to %s completed", f.mp4FilePath)
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
