package queues

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/syumai/workers/internal/jsutil"
)

type ConsumerMessage struct {
	// instance - The underlying instance of the JS message object passed by the cloudflare
	instance js.Value

	Id        string
	Timestamp time.Time
	Body      js.Value
	Attempts  int
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

func (m *ConsumerMessage) Ack() {
	m.instance.Call("ack")
}

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

type ConsumerMessageBatch struct {
	// instance - The underlying instance of the JS message object passed by the cloudflare
	instance js.Value

	Queue    string
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

func (b *ConsumerMessageBatch) AckAll() {
	b.instance.Call("ackAll")
}

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

func WithRetryDelay(d time.Duration) RetryOption {
	return func(o *retryOptions) {
		o.delaySeconds = int(d.Seconds())
	}
}
