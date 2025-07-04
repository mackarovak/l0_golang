package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:29092"})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	topic := "orders"

	message := `{"order_uid":"b563feb7b2b84b6test", "track_number":"WBILMTESTTRACK", "entry":"WBIL"}` // упрощённо

	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	if err != nil {
		fmt.Println("Failed to produce:", err)
	} else {
		fmt.Println("Message sent to Kafka")
	}
}
