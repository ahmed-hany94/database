package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Job struct {
	data string
}

type Event struct {
	content string
}

func wal(ctx context.Context, queue chan Job) {
	for {
		select {
		case <-ctx.Done():
			fmt.Print("received shutdown!!\n")
			return
		case job := <-queue:
			f, err := os.OpenFile("log", os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return
			}
			newEvent := &Event{content: job.data}
			f.Write([]byte(newEvent.content + "\n"))
			f.Sync()
		}
	}
}

func TestWorkerGracefulShutdown(t *testing.T) {
	os.Remove("log")
	ctx, cancel := context.WithCancel(context.Background())

	queue := make(chan Job, 10)

	go wal(ctx, queue)

	queue <- Job{data: "first write-ahead log item"}
	queue <- Job{data: "second write-ahead log item"}

	time.Sleep(100 * time.Millisecond)

	cancel()

	time.Sleep(10 * time.Millisecond)

	data, err := os.ReadFile("log")
	if err != nil {
		fmt.Printf("Error reading log file for asserts: %v", err)
	}
	dataStr := strings.Split(string(data), "\n")
	assert.Equal(t, "first write-ahead log item", dataStr[0])
	assert.Equal(t, "second write-ahead log item", dataStr[1])
}
