package cmd

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/jzhang046/croned-twitcasting-recorder/record"
	"github.com/jzhang046/croned-twitcasting-recorder/sink"
	"github.com/jzhang046/croned-twitcasting-recorder/twitcasting"
)

const (
	DirectRecordCmdName       = "direct"
	defaultRetryBackoffPeriod = 15 * time.Second
)

func RecordDirect(args []string) {
	log.Printf("Starting in recoding mode [%s] with PID [%d].. \n", DirectRecordCmdName, os.Getpid())

	directRecordCmd := flag.NewFlagSet(DirectRecordCmdName, flag.ExitOnError)
	streamer := directRecordCmd.String("streamer", "", "[required] streamer URL")
	retries := directRecordCmd.Int(
		"retries",
		0,
		"[optional] number of retries (default 0)", //default will not be auto appended for 0 value
	)
	retryBackoffPeriod := directRecordCmd.Duration(
		"retry-backoff",
		defaultRetryBackoffPeriod,
		"[optional] retry backoff period",
	)
	encodeOption := directRecordCmd.String("encode-option", "", "[optional] encode option of ffmpeg")

	directRecordCmd.Parse(args)

	if *streamer == "" {
		log.Println("Please provide a valid streamer URL ")
		directRecordCmd.Usage()
		os.Exit(1)
	}
	if *retries < 0 {
		log.Printf("number of retries must be non-negative ")
		directRecordCmd.Usage()
		os.Exit(1)
	}

	interruptCtx, afterGracefulInterrupt := newInterruptableCtx()

	for ; *retries >= 0; *retries-- {
		log.Printf(
			"Recording streamer [%s] with [%d] retries left and [%s] backoff \n",
			*streamer, *retries, *retryBackoffPeriod,
		)
		record.ToRecordFunc(&record.RecordConfig{
			Streamer:         *streamer,
			StreamUrlFetcher: twitcasting.GetWSStreamUrl,
			SinkProvider:     sink.NewFileSink,
			StreamRecorder:   twitcasting.RecordWS,
			RootContext:      interruptCtx,
			EncodeOption:     encodeOption,
		})()
		select {
		// wait for either interrupted or retry backoff period
		case <-interruptCtx.Done():
			<-afterGracefulInterrupt
			log.Fatal("Terminated on user interrupt")
		case <-time.After(*retryBackoffPeriod):
		}
	}
	log.Println("Recording all finished")
}
