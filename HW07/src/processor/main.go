package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	sqsClient *sqs.Client
	queueURL  string
	processed atomic.Int64
	firstMsg  atomic.Int64 // unix nanos of the first message received
)

func main() {
	queueURL = os.Getenv("SQS_QUEUE_URL")
	workers, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil || workers < 1 {
		workers = 1
	}
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	sqsClient = sqs.NewFromConfig(cfg)

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	slots := make(chan struct{}, workers)
	go logStats(workers, slots)
	runProcessor(context.Background(), workers, slots)
}

func runProcessor(ctx context.Context, workers int, slots chan struct{}) {
	for {
		free := workers - len(slots)
		if free == 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		out, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: int32(min(10, free)),
			WaitTimeSeconds:     20,
		})
		if err != nil {
			log.Println("receive failed:", err)
			time.Sleep(time.Second)
			continue
		}
		if len(out.Messages) > 0 {
			firstMsg.CompareAndSwap(0, time.Now().UnixNano())
		}
		for _, m := range out.Messages {
			slots <- struct{}{}
			go func(m types.Message) {
				defer func() { <-slots }()
				time.Sleep(3 * time.Second) // payment verification
				_, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(queueURL),
					ReceiptHandle: m.ReceiptHandle,
				})
				if err != nil {
					log.Println("delete failed:", err)
					return
				}
				processed.Add(1)
			}(m)
		}
	}
}

func logStats(workers int, slots chan struct{}) {
	var last int64
	for range time.Tick(10 * time.Second) {
		p := processed.Load()
		avg := 0.0
		if f := firstMsg.Load(); f > 0 {
			avg = float64(p) / time.Since(time.Unix(0, f)).Seconds()
		}
		log.Printf("workers=%d processed=%d in_flight=%d rate_10s=%.2f/s avg_rate=%.2f/s",
			workers, p, len(slots), float64(p-last)/10, avg)
		last = p
	}
}
