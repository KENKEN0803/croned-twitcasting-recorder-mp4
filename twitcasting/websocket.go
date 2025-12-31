package twitcasting

import (
	"errors"
	"log"
	"time"

	"github.com/sacOO7/gowebsocket"

	"github.com/jzhang046/croned-twitcasting-recorder-mp4/record"
	"github.com/jzhang046/croned-twitcasting-recorder-mp4/types"
)

func RecordWS(recordCtx record.RecordContext, streamInfo *types.StreamInfo, sinkChan chan<- []byte, cookie string) error {
	socket := gowebsocket.New(streamInfo.Url)
	defer func() {
		if socket.IsConnected {
			socket.Close()
		}
	}()
	defer close(sinkChan)

	streamer := recordCtx.GetStreamer()
	connectionResultChan := make(chan error, 1)

	socket.RequestHeader.Set("Origin", baseDomain)
	socket.RequestHeader.Set("User-Agent", userAgent)
	if streamInfo.Password != "" {
		socket.RequestHeader.Set("Sec-WebSocket-Protocol", streamInfo.Password)
	}
	// Only add cookie header if a cookie is provided (on retry)
	if cookie != "" {
		socket.RequestHeader.Set("Cookie", cookie)
	}

	socket.OnConnectError = func(err error, s gowebsocket.Socket) {
		connectionResultChan <- err
		close(connectionResultChan)
	}
	socket.OnConnected = func(s gowebsocket.Socket) {
		log.Printf("Connected to live stream for [%s], recording start \n", streamer)
		close(connectionResultChan) // Signal success
	}
	socket.OnTextMessage = func(message string, s gowebsocket.Socket) {
		log.Println("Received message", message)
	}
	socket.OnBinaryMessage = func(data []byte, s gowebsocket.Socket) {
		sinkChan <- data
	}
	socket.OnDisconnected = func(err error, s gowebsocket.Socket) {
		log.Printf("Disconnected from live stream of [%s] \n", streamer)
		recordCtx.Cancel()
	}

	socket.Connect()

	// Block until connection is established or fails
	select {
	case err := <-connectionResultChan:
		if err != nil {
			log.Printf("Connection failed for streamer [%s]: %v", streamer, err)
			return err // Return the error to the caller
		}
	case <-time.After(10 * time.Second): // Connection timeout
		err := errors.New("websocket connection timed out")
		log.Printf("Connection failed for streamer [%s]: %v", streamer, err)
		return err
	}

	// Connection successful, wait for recording to finish
	<-recordCtx.Done()
	log.Printf("Recording finished for streamer [%s].", streamer)

	return nil
}
