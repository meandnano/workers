package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/queues"
)

const queueName = "QUEUE"

func handleErr(w http.ResponseWriter, msg string, err error) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func main() {
	// start Qeueue consumer.
	// If we would not have an HTTP handler in this worker, we would use queues.Consume instead
	queues.ConsumeNonBlocking(consumeBatch)

	// start HTTP server
	http.HandleFunc("/", handleProduce)
	workers.Serve(nil)
}

func handleProduce(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer req.Body.Close()

	q, err := queues.NewProducer(queueName)
	if err != nil {
		handleErr(w, "failed to init queue", err)
	}

	contentType := req.Header.Get("Content-Type")
	switch contentType {
	case "text/plain":
		log.Println("Handling text content type")
		err = produceText(q, req)
	case "application/json":
		log.Println("Handling json content type")
		err = produceJson(q, req)
	default:
		log.Println("Handling bytes content type")
		err = produceBytes(q, req)
	}

	if err != nil {
		handleErr(w, "failed to handle request", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("message sent\n"))
}

func produceText(q *queues.Producer, req *http.Request) error {
	content, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	if len(content) == 0 {
		return fmt.Errorf("empty request body")
	}

	// text content type supports string and []byte messages
	if err := q.Send(content, queues.WithContentType(queues.QueueContentTypeText)); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func produceJson(q *queues.Producer, req *http.Request) error {
	var data any
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// json content type is default and therefore can be omitted
	// json content type supports messages of types that can be serialized to json
	if err := q.Send(data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func produceBytes(q *queues.Producer, req *http.Request) error {
	// bytes content type support messages of type []byte, string, and io.Reader
	if err := q.Send(req.Body, queues.WithContentType(queues.QueueContentTypeBytes)); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func consumeBatch(batch *queues.ConsumerMessageBatch) error {
	for _, msg := range batch.Messages {
		log.Printf("Received message: %v\n", msg.Body.Get("name").String())
	}

	batch.AckAll()
	return nil
}
