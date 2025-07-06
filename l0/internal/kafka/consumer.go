package kafka

import (
	"context"
	"encoding/json"
	"log"
	"l0/internal/cache"
	"l0/internal/models"
	"l0/internal/storage"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func StartConsumer(c *cache.Cache) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:29092",
		"group.id":          "order-consumers",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer consumer.Close()

	if err := consumer.Subscribe("orders", nil); err != nil {
		log.Fatal("Failed to subscribe to topic:", err)
	}

	log.Println("Kafka consumer started")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Consumer error: %v\n", err)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("Failed to unmarshal order: %v\n", err)
			continue
		}

		log.Printf("Processing order: %s\n", order.OrderUID)

		// Сохраняем в кэш
		c.Set(order)

		// Сохраняем в БД
		if err := storage.SaveOrder(context.Background(), order); err != nil {
			log.Printf("Failed to save order to DB: %v\n", err)
			continue
		}

		log.Printf("Order processed successfully: %s\n", order.OrderUID)
	}
}