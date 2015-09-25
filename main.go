package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ian-kent/gofigure"
	"github.com/shopify/sarama"
)

type config struct {
	gofigure interface{} `order:"env"`
	Brokers  string      `env:"KTEE_BROKERS"`
	OutTopic string      `env:"KTEE_OUT_TOPIC"`
	ErrTopic string      `env:"KTEE_ERR_TOPIC"`
}

func main() {
	var cmd *exec.Cmd
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigs:
				cmd.Process.Signal(sig)
			}
		}
	}()

	var cfg config
	if err := gofigure.Gofigure(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, "unexpected error configuring ktee")
		os.Exit(1)
	}

	var err error
	var producer sarama.SyncProducer

	if len(cfg.Brokers) > 0 {
		brokers := strings.Split(cfg.Brokers, ",")
		producer, err = sarama.NewSyncProducer(brokers, sarama.NewConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error connecting to Kafka brokers: %s\n", err)
			os.Exit(1)
		}

		defer func() {
			producer.Close()
		}()
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ktee args")
		os.Exit(1)
	}

	kwOut := kafkaWriter{producer, os.Stdout, cfg.OutTopic, new(bytes.Buffer), make(chan sarama.ProducerMessage)}
	kwErr := kafkaWriter{producer, os.Stderr, cfg.ErrTopic, new(bytes.Buffer), make(chan sarama.ProducerMessage)}

	defer func() {
		kwOut.Flush()
		kwErr.Flush()
	}()

	cmd = exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = kwOut
	cmd.Stderr = kwErr
	cmd.Env = os.Environ()

	err = cmd.Run()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			fmt.Fprintf(os.Stderr, "non-zero exit code: %s\n", err)
			if status, ok := err.(*exec.ExitError).Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stderr, "error executing command: %s\n", err)
			os.Exit(1)
		}
	}
}
