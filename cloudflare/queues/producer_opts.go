package queues

import (
	"syscall/js"
	"time"

	"github.com/syumai/workers/internal/jsutil"
)

type sendOptions struct {
	// ContentType - Content type of the message
	// Default is "json"
	ContentType contentType

	// DelaySeconds - The number of seconds to delay the message.
	// Default is 0
	DelaySeconds int
}

func (o *sendOptions) toJS() js.Value {
	obj := jsutil.NewObject()
	obj.Set("contentType", string(o.ContentType))

	if o.DelaySeconds != 0 {
		obj.Set("delaySeconds", o.DelaySeconds)
	}

	return obj
}

type SendOption func(*sendOptions)

// WithDelay changes the number of seconds to delay the message.
func WithDelay(d time.Duration) SendOption {
	return func(o *sendOptions) {
		o.DelaySeconds = int(d.Seconds())
	}
}

type batchSendOptions struct {
	// DelaySeconds - The number of seconds to delay the message.
	// Default is 0
	DelaySeconds int
}

func (o *batchSendOptions) toJS() js.Value {
	if o == nil {
		return js.Undefined()
	}

	obj := jsutil.NewObject()
	if o.DelaySeconds != 0 {
		obj.Set("delaySeconds", o.DelaySeconds)
	}

	return obj
}

type BatchSendOption func(*batchSendOptions)

// WithBatchDelay changes the number of seconds to delay the message.
func WithBatchDelay(d time.Duration) BatchSendOption {
	return func(o *batchSendOptions) {
		o.DelaySeconds = int(d.Seconds())
	}
}
