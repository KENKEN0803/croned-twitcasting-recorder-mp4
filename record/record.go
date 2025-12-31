package record

import (
	"context"
	"log"
	"strings"

	"github.com/jzhang046/croned-twitcasting-recorder-mp4/config"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/sink" // Add sink import
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/types"
)

type RecordConfig struct {
	Streamer         string
	StreamUrlFetcher func(streamer, cookie string) (*types.StreamInfo, error)
	SinkProvider     func(RecordContext) (chan<- []byte, string, error) // Updated signature
	StreamRecorder   func(recordCtx RecordContext, streamInfo *types.StreamInfo, sinkChan chan<- []byte, cookie string) error
	RootContext      context.Context
	EncodeOption     *string
	AppConfig        *config.Config
}

func ToRecordFunc(recordConfig *RecordConfig) func() {
	streamer := recordConfig.Streamer
	return func() {
		var cookie string
		if recordConfig.AppConfig != nil && recordConfig.AppConfig.Twitcasting != nil {
			cookie = recordConfig.AppConfig.Twitcasting.Cookie
		}

		// First attempt, without cookie
		streamInfo, err := recordConfig.StreamUrlFetcher(streamer, "")
		if err != nil {
			//log.Printf("Error fetching stream info for streamer [%s]: %v\n", streamer, err)
			return
		}

		// Prepare for recording
		streamTitle, err := GetStreamTitle(streamer)
		if err != nil {
			log.Printf("Error fetching stream title for streamer [%s]: %v\n", streamer, err)
		}
		log.Printf("Stream Title is %s\n", streamTitle)

		recordCtx := newRecordContext(recordConfig.RootContext, streamer, streamInfo.Url, streamTitle, recordConfig.EncodeOption, streamInfo.IsMembershipStream)
		sinkChan, tsFilePath, err := recordConfig.SinkProvider(recordCtx) // Capture tsFilePath
		if err != nil {
			log.Println("Error creating recording file: ", err)
			return
		}

		// Attempt to record
		err = recordConfig.StreamRecorder(recordCtx, streamInfo, sinkChan, "")
		if err != nil && strings.Contains(err.Error(), "bad handshake") && cookie != "" {
			log.Printf("Authentication error for streamer [%s]. Retrying with cookie.", streamer)
			// Delete the empty file before retry
			_ = sink.RemoveFileIfSmall(tsFilePath, 1024) // Delete the file if small
			recordCtx.Cancel()                           // Cancel the previous context

			// Fetch new stream info with cookie
			streamInfo, err = recordConfig.StreamUrlFetcher(streamer, cookie)
			if err != nil {
				log.Printf("Error fetching stream info with cookie for streamer [%s]: %v\n", streamer, err)
				return
			}
			log.Printf("Fetched new stream URL for streamer [%s]: %s. ", streamer, streamInfo.Url)

			// Create new context and sink
			recordCtx = newRecordContext(recordConfig.RootContext, streamer, streamInfo.Url, streamTitle, recordConfig.EncodeOption, streamInfo.IsMembershipStream)
			var retryTsFilePath string
			sinkChan, retryTsFilePath, err = recordConfig.SinkProvider(recordCtx) // Capture tsFilePath for retry
			if err != nil {
				log.Println("Error creating recording file for retry: ", err)
				return
			}
			// Retry recording
			err = recordConfig.StreamRecorder(recordCtx, streamInfo, sinkChan, cookie)
			if err != nil {
				log.Printf("Recording retry failed for streamer [%s]: %v", streamer, err)
				if strings.Contains(err.Error(), "bad handshake") {
					_ = sink.RemoveFileIfSmall(retryTsFilePath, 1024) // Delete file if retry also fails handshake and file is small
				}
			}

		} else if err != nil {
			log.Printf("Recording failed for streamer [%s]: %v", streamer, err)
			// Also delete the file if recording failed for other reasons (not a handshake retry) and file is small
			_ = sink.RemoveFileIfSmall(tsFilePath, 1024)
		}
	}
}
