package application

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumerHandler func([]byte)
type KafkaConsumerHandlers map[string]KafkaConsumerHandler

type Kafka struct {
	Writer    *kafka.Writer
	consumers kafkaConsumers
}

type kafkaConsumers struct {
	readers []*kafka.Reader
	wg      *sync.WaitGroup
	cancel  context.CancelFunc
}

const kafkaAddr = "localhost:9092"

func StartKafka(topicsToCreate []string, consumerGroup string, handlers KafkaConsumerHandlers) Kafka {
	kafkaCreateTopics(topicsToCreate)
	consumers := kafkaStartConsumers(consumerGroup, handlers)
	writer := &kafka.Writer{Addr: kafka.TCP(kafkaAddr), Async: true}

	return Kafka{Writer: writer, consumers: consumers}
}

func ShutdownKafka(k Kafka) {
	k.consumers.cancel()
	log.Println("Requested Kafka consumers to finish")
	k.consumers.wg.Wait()
	log.Println("All Kafka consumers exited handling loops")

	for _, reader := range k.consumers.readers {
		if err := reader.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}
	log.Println("All Kafka consumers closed!")

	if err := k.Writer.Close(); err == nil {
		log.Println("Kafka writer successfully closed!")
	} else {
		log.Printf("Failed to close Kafka writer: %v", err)
	}
}

func kafkaCreateTopics(topicsToCreate []string) {
	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		log.Fatalf("Failed to establish Kafka connection: %v", err)
	}
	defer conn.Close()

	// Create needed topics
	topicConfigs := make([]kafka.TopicConfig, 0, len(topicsToCreate))

	for _, topicName := range topicsToCreate {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	if err := conn.CreateTopics(topicConfigs...); err != nil {
		log.Fatalf("Failed to create Kafka topics: %v", err)
	}

	// List topics in cluster
	partitions, err := conn.ReadPartitions()
	if err != nil {
		log.Fatalf("Failed to read Kafka partitions: %v", err)
	}

	topicsSet := map[string]struct{}{}

	for _, p := range partitions {
		topicsSet[p.Topic] = struct{}{}
	}

	topics := make([]string, 0, len(topicsSet))
	for k := range topicsSet {
		topics = append(topics, k)
	}
	log.Printf("Kafka initialized, topics in cluster: %v", topics)
}

func kafkaStartConsumers(consumerGroup string, handlers KafkaConsumerHandlers) kafkaConsumers {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	readers := make([]*kafka.Reader, 0, len(handlers))
	for topic, handler := range handlers {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{kafkaAddr},
			GroupID:        consumerGroup,
			Topic:          topic,
			CommitInterval: time.Second,
		})
		wg.Add(1)
		go kafkaStartConsumer(ctx, reader, handler, wg)
	}

	return kafkaConsumers{readers: readers, wg: wg, cancel: cancel}
}

func kafkaStartConsumer(ctx context.Context, reader *kafka.Reader, handler KafkaConsumerHandler, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if err != context.Canceled {
				log.Printf("Failed to consume message from Kafka: %v", err)
			}
			break
		}

		handler(message.Value)
	}
}
