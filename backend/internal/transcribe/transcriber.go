package transcribe

import "context"

type Transcriber interface {
	Transcribe(ctx context.Context, audioPath string) (string, error)
}
