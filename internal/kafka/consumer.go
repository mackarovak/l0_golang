package kafka

import (
 "l0/internal/cache"
 "l0/internal/models"
 "context"
 "encoding/json"
 "log"
 "l0/internal/storage"

 "github.com/confluentinc/confluent-kafka-go/kafka"
)

func StartConsumer(c *cache.Cache) {
 conf := &kafka.ConfigMap{
  "bootstrap.servers": "localhost:29092",
  "group.id":          "order-consumer",
  "auto.offset.reset": "earliest",
 }

 consumer, err := kafka.NewConsumer(conf)
 if err != nil {
  log.Fatal(err)
 }

 consumer.Subscribe("orders", nil)

 for {
  msg, err := consumer.ReadMessage(-1)
  if err != nil {
   log.Println("Consumer error:", err)
   continue
  }

  var order models.Order
  if err := json.Unmarshal(msg.Value, &order); err != nil {
   log.Println("Bad JSON:", err)
   continue
  }

  log.Println("New order:", order.OrderUID)
  c.Set(order)

  if err := storage.SaveOrder(context.Background(), order); err != nil {
	log.Println("failed to save to db:", err)
}

 }
}