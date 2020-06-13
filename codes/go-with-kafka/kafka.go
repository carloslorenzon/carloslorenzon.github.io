package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	topic         = "meu-topico"
	groupConsumer = "meu-grupo"
	broker        = "localhost:9092"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go [producer|consumer]")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "producer":
		runProducer()
	case "consumer":
		runConsumer()
	default:
		fmt.Println("Modo não reconhecido. Use 'producer' ou 'consumer'")
	}
}

func runConsumer() {
	topics := []string{topic}
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        broker,
		"broker.address.family":    "v4",
		"group.id":                 groupConsumer,
		"session.timeout.ms":       6000,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe to topics: %s\n", err)
		os.Exit(1)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}

				//Criar um TopicPartition com o offset específico
				tp := kafka.TopicPartition{
					Topic:     e.TopicPartition.Topic,
					Partition: e.TopicPartition.Partition,
					Offset:    e.TopicPartition.Offset + 1, // Offset para ser commitado (próxima mensagem a ser lida)
				}

				// Commit manual do offset específico
				offsets, err := c.CommitOffsets([]kafka.TopicPartition{tp})
				if err != nil {
					fmt.Printf("Failed to commit offsets: %s\n", err)
				} else {
					fmt.Printf("Committed offsets: %v\n", offsets)
				}

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}

func runProducer() {
	config := &kafka.ConfigMap{
		"bootstrap.servers": broker,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	defer producer.Close()

	topicName := topic
	for _, word := range []string{"Hello", "Kafka", "World"} {
		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}

		err := producer.Produce(message, nil)
		if err != nil {
			fmt.Printf("Failed to produce message: %s\n", err)
			continue
		}

		e := <-producer.Events()
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			fmt.Printf("Delivered message to %v\n", m.TopicPartition)
		}
	}

	// Flush and wait for all messages to be delivered
	producer.Flush(15 * 1000)
}
