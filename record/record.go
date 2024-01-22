package record

import (
	"context"
	"log"
)

type RecordConfig struct {
	Streamer         string
	StreamUrlFetcher func(string) (string, error)
	SinkProvider     func(RecordContext) (chan<- []byte, error)
	StreamRecorder   func(RecordContext, chan<- []byte)
	RootContext      context.Context
	EncodeOption     *string
}

func ToRecordFunc(recordConfig *RecordConfig) func() {
	streamer := recordConfig.Streamer
	return func() {
		streamUrl, err := recordConfig.StreamUrlFetcher(streamer)
		if err != nil {
			log.Printf("Error fetching stream URL for streamer [%s]: %v\n", streamer, err)
			return
		}
		log.Printf("Fetched stream URL for streamer [%s]: %s. ", streamer, streamUrl)

		streamTitle, err := GetStreamTitle(streamer)
		if err != nil {
			log.Printf("Error fetching stream title for streamer [%s]: %v\n", streamer, err)
		}
		log.Printf("Stream Title is %s\n", streamTitle)

		recordCtx := newRecordContext(recordConfig.RootContext, streamer, streamUrl, streamTitle, recordConfig.EncodeOption)

		sinkChan, err := recordConfig.SinkProvider(recordCtx)
		if err != nil {
			log.Println("Error creating recording file: ", err)
			return
		}

		recordConfig.StreamRecorder(recordCtx, sinkChan)
	}
}
