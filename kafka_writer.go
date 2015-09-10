package main

import (
	"bytes"
	"io"

	"github.com/shopify/sarama"
)

type kafkaWriter struct {
	producer sarama.SyncProducer
	writer   io.Writer
	topic    string
	buffer   *bytes.Buffer
}

func (w kafkaWriter) send() error {
	for {
		ln, err := w.buffer.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// TODO: handle these errors?
			break
		}

		message := &sarama.ProducerMessage{
			Topic: w.topic,
			Value: sarama.ByteEncoder(ln),
		}

		if _, _, err := w.producer.SendMessage(message); err != nil {
			// TODO: handle errors, buffer, etc
			w.writer.(io.Closer).Close()
			return err
		}

		w.writer.Write(ln)
	}

	return nil
}

func (w kafkaWriter) Flush() error {
	return w.send()
}

func (w kafkaWriter) Write(b []byte) (n int, err error) {
	// TODO: support optional in-memory buffering, memory-mapped files and file buffering

	if w.producer != nil && len(w.topic) > 0 {
		n, err = w.buffer.Write(b)
		if err != nil {
			return
		}

		err = w.send()
		return
	}

	return w.writer.Write(b)
}
