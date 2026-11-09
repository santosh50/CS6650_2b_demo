package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending, processing, completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// One slot = the payment processor handles one order at a time.
var paymentSlots = make(chan struct{}, 1)

var (
	snsClient *sns.Client
	topicArn  string
)

func main() {
	topicArn = os.Getenv("SNS_TOPIC_ARN")
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	snsClient = sns.NewFromConfig(cfg)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/orders/sync", createOrderSync)
	r.POST("/orders/async", createOrderAsync)
	log.Fatal(r.Run(":8080"))
}

func verifyPayment(o *Order) {
	paymentSlots <- struct{}{}
	defer func() { <-paymentSlots }()
	time.Sleep(3 * time.Second)
	o.Status = "completed"
}

func newOrder(c *gin.Context) (*Order, bool) {
	var o Order
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, false
	}
	o.OrderID = uuid.NewString()
	o.CreatedAt = time.Now()
	return &o, true
}

func createOrderSync(c *gin.Context) {
	o, ok := newOrder(c)
	if !ok {
		return
	}
	o.Status = "processing"
	verifyPayment(o)
	c.JSON(http.StatusOK, o)
}

func createOrderAsync(c *gin.Context) {
	o, ok := newOrder(c)
	if !ok {
		return
	}
	o.Status = "pending"
	body, _ := json.Marshal(o)
	_, err := snsClient.Publish(c.Request.Context(), &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(body)),
	})
	if err != nil {
		log.Println("publish failed:", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not queue order"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"order_id": o.OrderID, "status": o.Status})
}
