package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sillkiw/wb-l0/internal/generator"
	"github.com/sillkiw/wb-l0/internal/kafka"

	"github.com/joho/godotenv"
	k "github.com/segmentio/kafka-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}
	broker := os.Getenv("KAFKA_EXTERNAL")
	topic := os.Getenv("KAFKA_TOPIC")
	countStr := os.Getenv("PRODUCER_COUNT")
	intervalStr := os.Getenv("PRODUCER_INTERVAL")

	log.Printf("Broker = %s", broker)
	log.Printf("topic = %s", topic)
	count, _ := strconv.Atoi(countStr)
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = time.Second
	}

	producer := kafka.NewProducer([]string{broker}, topic)
	defer producer.Close()

	if err := checkKafkaConnection(broker); err != nil {
		log.Printf("failed to connect %s", err)
	}

	i := 0
	for {
		order := generator.NewFakeOrder(i)
		data, _ := json.Marshal(order)

		if err := producer.Send([]byte(order.OrderUID), data); err != nil {
			log.Printf("failed: %v", err)
		} else {
			log.Printf("sent order %s", order.OrderUID)
		}

		time.Sleep(interval)
		i++
		if count > 0 && i >= count {
			break
		}
	}
}

func checkKafkaConnection(broker string) error {
	conn, err := k.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	// пробуем получить список топиков
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	fmt.Printf("Kafka connected, %d partitions found\n", len(partitions))
	return nil
}
