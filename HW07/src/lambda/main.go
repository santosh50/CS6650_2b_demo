package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Order struct {
	OrderID    string `json:"order_id"`
	CustomerID int    `json:"customer_id"`
}

func handler(ctx context.Context, e events.SNSEvent) error {
	for _, r := range e.Records {
		var o Order
		if err := json.Unmarshal([]byte(r.SNS.Message), &o); err != nil {
			return err
		}
		time.Sleep(3 * time.Second) // payment verification
		log.Printf("processed order %s", o.OrderID)
	}
	return nil
}

func main() { lambda.Start(handler) }
