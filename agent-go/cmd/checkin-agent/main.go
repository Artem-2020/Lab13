package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"hotel-agent/internal/checkin"
)

const subject = "hotel.tasks.checkin_checkout"
const queueGroup = "hotel_agents"

var processedTasks uint64

func main() {
	logger, logFile := newLogger("logs/agent.log")
	if logFile != nil {
		defer logFile.Close()
	}

	natsURL := env("NATS_URL", nats.DefaultURL)
	agentID := env("AGENT_ID", "checkin-agent-local")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Fatalf("ERROR connect to NATS: %v", err)
	}
	defer nc.Close()

	_, err = nc.QueueSubscribe(subject, queueGroup, func(msg *nats.Msg) {
		var task checkin.Task
		if err := json.Unmarshal(msg.Data, &task); err != nil {
			logger.Printf("ERROR invalid JSON: %v", err)
			return
		}

		result := checkin.Process(task, agentID)
		payload, err := json.Marshal(result)
		if err != nil {
			logger.Printf("ERROR marshal result for task %s: %v", task.TaskID, err)
			return
		}

		if task.ReplyTo == "" {
			logger.Printf("ERROR task %s has empty reply_to", task.TaskID)
			return
		}

		if err := nc.Publish(task.ReplyTo, payload); err != nil {
			logger.Printf("ERROR publish result for task %s: %v", task.TaskID, err)
			return
		}

		count := atomic.AddUint64(&processedTasks, 1)
		logger.Printf("INFO processed task_id=%s action=%s total=%d", task.TaskID, task.Action, count)
	})
	if err != nil {
		logger.Fatalf("ERROR subscribe: %v", err)
	}

	logger.Printf("INFO %s subscribed to %s in queue group %s", agentID, subject, queueGroup)
	select {}
}

func newLogger(path string) (*log.Logger, *os.File) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	return log.New(io.MultiWriter(os.Stdout, file), "", log.LstdFlags), file
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
