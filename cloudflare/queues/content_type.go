package queues

import (
	"fmt"
	"io"
	"syscall/js"

	"github.com/syumai/workers/internal/jsutil"
)

type QueueContentType string

const (
	QueueContentTypeJSON  QueueContentType = "json"
	QueueContentTypeText  QueueContentType = "text"
	QueueContentTypeBytes QueueContentType = "bytes"
	QueueContentTypeV8    QueueContentType = "v8"
)

func (o QueueContentType) mapValue(val any) (js.Value, error) {
	switch o {
	case QueueContentTypeText:
		switch v := val.(type) {
		case string:
			return js.ValueOf(v), nil
		case []byte:
			return js.ValueOf(string(v)), nil
		default:
			return js.Undefined(), fmt.Errorf("invalid value type for text content type: %T", val)
		}

	case QueueContentTypeBytes:
		var b []byte
		switch v := val.(type) {
		case string:
			b = []byte(v)
		case []byte:
			b = v
		case io.Reader:
			var err error
			b, err = io.ReadAll(v)
			if err != nil {
				return js.Undefined(), fmt.Errorf("failed to read bytes from reader: %w", err)
			}
		default:
			return js.Undefined(), fmt.Errorf("invalid value type for bytes content type: %T", val)
		}

		ua := jsutil.NewUint8Array(len(b))
		js.CopyBytesToJS(ua, b)
		return ua.Get("buffer"), nil

	case QueueContentTypeJSON, QueueContentTypeV8:
		return js.ValueOf(val), nil
	}

	return js.Undefined(), fmt.Errorf("unknown content type: %s", o)
}
