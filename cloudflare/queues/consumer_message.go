package queues

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/syumai/workers/internal/jsutil"
)

// ConsumerMessage represents a message of the batch received by the consumer.
//   - https://developers.cloudflare.com/queues/configuration/javascript-apis/#message
type ConsumerMessage struct {
	// instance - The underlying instance of the JS message object passed by the cloudflare
	instance js.Value

	// Id - The unique Cloudflare-generated identifier of the message
	Id string
	// Timestamp - The time when the message was enqueued
	Timestamp time.Time
	// Body - The message body. Could be accessed directly or using converting helpers as StringBody, BytesBody, IntBody, FloatBody.
	Body js.Value
	// Attempts - The number of times the message delivery has been retried.
	Attempts int
}

func newConsumerMessage(obj js.Value) (*ConsumerMessage, error) {
	timestamp, err := jsutil.DateToTime(obj.Get("timestamp"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse message timestamp: %v", err)
	}

	return &ConsumerMessage{
		instance:  obj,
		Id:        obj.Get("id").String(),
		Body:      obj.Get("body"),
		Attempts:  obj.Get("attempts").Int(),
		Timestamp: timestamp,
	}, nil
}

// Ack acknowledges the message as successfully delivered despite the result returned from the consuming function.
//   - https://developers.cloudflare.com/queues/configuration/javascript-apis/#message
func (m *ConsumerMessage) Ack() {
	m.instance.Call("ack")
}

// Retry marks the message to be re-delivered.
// The message will be retried after the optional delay configured with RetryOption.
func (m *ConsumerMessage) Retry(opts ...RetryOption) {
	var o *retryOptions
	if len(opts) > 0 {
		o = &retryOptions{}
		for _, opt := range opts {
			opt(o)
		}
	}

	m.instance.Call("retry", o.toJS())
}

func (m *ConsumerMessage) StringBody() (string, error) {
	if m.Body.Type() != js.TypeString {
		return "", fmt.Errorf("message body is not a string: %v", m.Body)
	}
	return m.Body.String(), nil
}

func (m *ConsumerMessage) BytesBody() ([]byte, error) {
	switch m.Body.Type() {
	case js.TypeString:
		return []byte(m.Body.String()), nil
	case js.TypeObject:
		if m.Body.InstanceOf(jsutil.Uint8ArrayClass) || m.Body.InstanceOf(jsutil.Uint8ClampedArrayClass) {
			b := make([]byte, m.Body.Get("byteLength").Int())
			js.CopyBytesToGo(b, m.Body)
			return b, nil
		}
	}

	return nil, fmt.Errorf("message body is not a byte array: %v", m.Body)
}

func (m *ConsumerMessage) IntBody() (int, error) {
	if m.Body.Type() == js.TypeNumber {
		return m.Body.Int(), nil
	}

	return 0, fmt.Errorf("message body is not a number: %v", m.Body)
}

func (m *ConsumerMessage) FloatBody() (float64, error) {
	if m.Body.Type() == js.TypeNumber {
		return m.Body.Float(), nil
	}

	return 0, fmt.Errorf("message body is not a number: %v", m.Body)
}

// ConsumerMessageBatch represents a batch of messages received by the consumer. The size of the batch is determined by the
// worker configuration.
//   - https://developers.cloudflare.com/queues/configuration/configure-queues/#consumer
//   - https://developers.cloudflare.com/queues/configuration/javascript-apis/#messagebatch
type ConsumerMessageBatch struct {
	// instance - The underlying instance of the JS message object passed by the cloudflare
	instance js.Value

	// Queue - The name of the queue from which the messages were received
	Queue string

	// Messages - The messages in the batch
	Messages []*ConsumerMessage
}

func newConsumerMessageBatch(obj js.Value) (*ConsumerMessageBatch, error) {
	msgArr := obj.Get("messages")
	messages := make([]*ConsumerMessage, msgArr.Length())
	for i := 0; i < msgArr.Length(); i++ {
		m, err := newConsumerMessage(msgArr.Index(i))
		if err != nil {
			return nil, fmt.Errorf("failed to parse message %d: %v", i, err)
		}
		messages[i] = m
	}

	return &ConsumerMessageBatch{
		instance: obj,
		Queue:    obj.Get("queue").String(),
		Messages: messages,
	}, nil
}

// AckAll acknowledges all messages in the batch as successfully delivered despite the result returned from the consuming function.
//   - https://developers.cloudflare.com/queues/configuration/javascript-apis/#messagebatch
func (b *ConsumerMessageBatch) AckAll() {
	b.instance.Call("ackAll")
}

// RetryAll marks all messages in the batch to be re-delivered.
// The messages will be retried after the optional delay configured with RetryOption.
//   - https://developers.cloudflare.com/queues/configuration/javascript-apis/#messagebatch
func (b *ConsumerMessageBatch) RetryAll(opts ...RetryOption) {
	var o *retryOptions
	if len(opts) > 0 {
		o = &retryOptions{}
		for _, opt := range opts {
			opt(o)
		}
	}

	b.instance.Call("retryAll", o.toJS())
}

type retryOptions struct {
	delaySeconds int
}

func (o *retryOptions) toJS() js.Value {
	if o == nil {
		return js.Undefined()
	}

	obj := jsutil.NewObject()
	if o.delaySeconds != 0 {
		obj.Set("delaySeconds", o.delaySeconds)
	}

	return obj
}

type RetryOption func(*retryOptions)

// WithRetryDelay sets the delay in seconds before the messages delivery is retried.
func WithRetryDelay(d time.Duration) RetryOption {
	return func(o *retryOptions) {
		o.delaySeconds = int(d.Seconds())
	}
}
