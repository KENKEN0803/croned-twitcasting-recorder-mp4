package record

import "context"

type RecordContext interface {
	// Done would be closed when work done.
	Done() <-chan struct{}

	// Err explains the reason when this context is Done().
	Err() error

	// Cancel cancels the record.
	Cancel()

	// GetStreamUrl returns the stream URL of this context.
	GetStreamUrl() string

	// GetStreamer returns streamer's screen ID of this context.
	GetStreamer() string

	// GetEncodeOption returns ffmpeg video encode options of this context
	GetEncodeOption() *string
}

type recordContextImpl struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type contextKey string

const (
	streamerKey     = contextKey("streamer")
	streamUrlKey    = contextKey("streamUrl")
	encodeOptionKey = contextKey("encodeOption")
)

func newRecordContext(ctx context.Context, streamer, streamUrl string, encodeOption *string) RecordContext {
	ctx, cancelFunc := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, streamUrlKey, streamUrl)
	ctx = context.WithValue(ctx, streamerKey, streamer)
	ctx = context.WithValue(ctx, encodeOptionKey, encodeOption)
	return &recordContextImpl{ctx, cancelFunc}
}

func (ctxImpl *recordContextImpl) Done() <-chan struct{} {
	return ctxImpl.ctx.Done()
}

func (ctxImpl *recordContextImpl) Err() error {
	return ctxImpl.ctx.Err()
}

func (ctxImpl *recordContextImpl) Cancel() {
	ctxImpl.cancelFunc()
}

func (ctxImpl *recordContextImpl) GetStreamUrl() string {
	return ctxImpl.ctx.Value(streamUrlKey).(string)
}

func (ctxImpl *recordContextImpl) GetStreamer() string {
	return ctxImpl.ctx.Value(streamerKey).(string)
}

func (ctxImpl *recordContextImpl) GetEncodeOption() *string {
	return ctxImpl.ctx.Value(encodeOptionKey).(*string)
}
