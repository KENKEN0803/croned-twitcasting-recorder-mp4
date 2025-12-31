package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jzhang046/croned-twitcasting-recorder-mp4/cmd"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/config"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/record"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/sink"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/uploader"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

var availableCmds = []string{cmd.CronedRecordCmdName, cmd.DirectRecordCmdName}

func main() {
	cfg := config.GetDefaultConfig()

	var defaultUploader uploader.Uploader
	var err error
	if cfg.R2 != nil && cfg.R2.Enabled {
		defaultUploader, err = uploader.NewR2Uploader(cfg.R2)
		if err != nil {
			log.Printf("Failed to initialize R2 defaultUploader: %v. Upload will be disabled.", err)
			defaultUploader = nil // Ensure defaultUploader is nil on error
		} else {
			log.Println("R2 defaultUploader initialized.")
		}
	}

	sinkProvider := func(recordCtx record.RecordContext) (chan<- []byte, error) {
		return sink.NewFileSink(recordCtx, defaultUploader)
	}

	if len(os.Args) < 2 {
		log.Println("Record mode not specified; supported modes:", availableCmds)
		cmd.RecordCroned(cfg, sinkProvider)
	} else {
		switch os.Args[1] {
		case cmd.CronedRecordCmdName:
			cmd.RecordCroned(cfg, sinkProvider)
		case cmd.DirectRecordCmdName:
			cmd.RecordDirect(cfg, os.Args[2:], sinkProvider)
		default:
			log.Fatalf(
				"Unknown record mode [%s]; supported modes: %s",
				os.Args[1],
				availableCmds,
			)
		}
	}
}
