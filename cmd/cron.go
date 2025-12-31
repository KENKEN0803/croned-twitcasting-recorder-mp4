package cmd

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/jzhang046/croned-twitcasting-recorder-mp4/config"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/record"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/twitcasting"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/types"
)

const CronedRecordCmdName = "croned"

func RecordCroned(cfg *config.Config, sinkProvider func(record.RecordContext) (chan<- []byte, error)) {
	log.Printf("Starting in recoding mode [%s] with PID [%d].. \n", CronedRecordCmdName, os.Getpid())

	if len(cfg.Streamers) == 0 {
		log.Println("No streamers configured in config.yaml for cron mode. Exiting.")
		return
	}

	c := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
		cron.SkipIfStillRunning(cron.DefaultLogger),
	))

	interruptCtx, afterGracefulInterrupt := newInterruptableCtx()

	for _, streamerConfig := range cfg.Streamers {
		originalJob := record.ToRecordFunc(&record.RecordConfig{
			Streamer: streamerConfig.ScreenId,
			StreamUrlFetcher: func(streamer, cookie string) (*types.StreamInfo, error) {
				return twitcasting.GetWSStreamUrl(streamer, cookie)
			},
			SinkProvider:   sinkProvider,
			StreamRecorder: twitcasting.RecordWS,
			RootContext:    interruptCtx,
			EncodeOption:   streamerConfig.EncodeOption,
			AppConfig:      cfg,
		})

		wrappedJob := func() {
			// delay for 1 to 10 seconds
			delay := time.Duration(rand.Intn(10)+1) * time.Second
			//log.Printf("Schedule triggered for streamer [%s], waiting for a random delay of %v before starting...", streamerConfig.ScreenId, delay)
			time.Sleep(delay)
			originalJob()
		}

		if _, err := c.AddFunc(
			streamerConfig.Schedule,
			wrappedJob,
		); err != nil {
			log.Fatalln("Failed adding record schedule: ", err)
		} else {
			log.Printf("Added schedule [%s] for streamer [%s] \n", streamerConfig.Schedule, streamerConfig.ScreenId)
		}
	}

	c.Start()
	log.Println("croned recorder started ")

	// interrupt => stop cron and wait for all task to complete => wait for graceful interrupt
	<-interruptCtx.Done()
	<-c.Stop().Done()
	<-afterGracefulInterrupt

	log.Fatal("Terminated on user interrupt")
}
