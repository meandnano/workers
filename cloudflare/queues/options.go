package queues

import (
	"syscall/js"
	"time"

	"github.com/syumai/workers/internal/jsutil"
)

type sendOpions struct {
	// ContentType - Content type of the message
	// Default is "json"
	ContentType QueueContentType

	// DelaySeconds - The number of seconds to delay the message.
	// Default is 0
	DelaySeconds int
}

func defaultSendOptions() *sendOpions {
	return &sendOpions{
		ContentType: QueueContentTypeJSON,
	}
}

func (o *sendOpions) toJS() js.Value {
	obj := jsutil.NewObject()
	obj.Set("contentType", string(QueueContentTypeJSON))

	if o.DelaySeconds != 0 {
		obj.Set("delaySeconds", o.DelaySeconds)
	}

	return obj
}

type SendOption func(*sendOpions)

// WithContentType changes the content type of the message.
func WithContentType(contentType QueueContentType) SendOption {
	return func(o *sendOpions) {
		o.ContentType = contentType
	}
}

// WithDelay changes the number of seconds to delay the message.
func (q *Queue) WithDelay(d time.Duration) SendOption {
	return func(o *sendOpions) {
		o.DelaySeconds = int(d.Seconds())
	}
}
