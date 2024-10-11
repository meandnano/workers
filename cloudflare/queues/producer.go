package queues

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/syumai/workers/internal/jsutil"
)

type Queue struct {
	// instance - Objects that Queue API belongs to. Default is Global
	instance js.Value
}

func NewQueue(varName string) (*Queue, error) {
	inst := js.Global().Get(varName)
	if inst.IsUndefined() {
		return nil, fmt.Errorf("%s is undefined", varName)
	}
	return &Queue{instance: inst}, nil
}

func (q *Queue) Send(content any, opts ...SendOption) error {
	if q.instance.IsUndefined() {
		return errors.New("queue object not found")
	}

	options := defaultSendOptions()
	for _, opt := range opts {
		opt(options)
	}

	jsValue, err := options.ContentType.mapValue(content)
	if err != nil {
		return err
	}

	p := q.instance.Call("send", jsValue, options.toJS())
	_, err = jsutil.AwaitPromise(p)
	return err
}
