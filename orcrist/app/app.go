package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type App struct {
	HTTPRoutes    map[string]http.HandlerFunc
	KafkaHandlers map[string]func([]byte)
}

var Redis *redis.Client
var Kafka *kafka.Writer
var kafkaConsumers []*kafka.Reader
var httpServer *http.Server

const KeepAliveTopic = "keep-alive"
const OrcTokensTopic = "orc-tokens"

func Start(app *App) {
	startRedis()
	startKafka(app)
	startHTTPServer(app)
}

func Shutdown() {
	shutdownHTTPServer()
	shutdownKafka()
	shutdownRedis()
}

func startHTTPServer(app *App) {
	httpServer = &http.Server{Addr: ":8080"}

	for route, handler := range app.HTTPRoutes {
		http.HandleFunc(route, handler)
	}

	go func() {
		log.Println("Starting HTTP server on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func shutdownHTTPServer() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err == nil {
		log.Println("HTTP server shutdown successful!")
	} else {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

func startRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := Redis.Ping(context.Background()).Err(); err == nil {
		log.Println("Redis initialized")
	} else {
		log.Fatalf("Redis initialization failed: %v", err)
	}
}

func shutdownRedis() {
	if err := Redis.Close(); err == nil {
		log.Println("Redis client successfully closed!")
	} else {
		log.Printf("Failed to close Redis client: %v", err)
	}
}

const kafkaAddr = "localhost:9092"
const kafkaConsumerGroup = "orcrist"

func startKafka(app *App) {
	kafkaCreateTopics()
	kafkaStartConsumers(app)
	Kafka = &kafka.Writer{Addr: kafka.TCP(kafkaAddr), Async: true}
}

func shutdownKafka() {
	kafkaConsumersCancel()
	log.Println("Requested Kafka consumers to finish")
	kafkaConsumersWG.Wait()
	log.Println("All Kafka consumers exited handling loops")

	for _, reader := range kafkaConsumers {
		if err := reader.Close(); err != nil {
			log.Printf("Failed to close Kafka consumer: %v", err)
		}
	}
	log.Println("All Kafka consumers closed!")

	if err := Kafka.Close(); err == nil {
		log.Println("Kafka writer successfully closed!")
	} else {
		log.Printf("Failed to close Kafka writer: %v", err)
	}
}

func kafkaCreateTopics() {
	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		log.Fatalf("Failed to establish Kafka connection: %v", err)
	}
	defer conn.Close()

	// Create needed topics
	topicsToCreate := [...]string{KeepAliveTopic, OrcTokensTopic}
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

var kafkaConsumersWG sync.WaitGroup
var kafkaConsumersCancel context.CancelFunc

func kafkaStartConsumers(app *App) {
	ctx, cancel := context.WithCancel(context.Background())
	kafkaConsumersCancel = cancel

	kafkaConsumers = make([]*kafka.Reader, 0, len(app.KafkaHandlers))
	for topic, handler := range app.KafkaHandlers {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{kafkaAddr},
			GroupID:        kafkaConsumerGroup,
			Topic:          topic,
			CommitInterval: time.Second,
		})
		kafkaConsumersWG.Add(1)
		go kafkaStartConsumer(reader, handler, ctx)
	}
}

func kafkaStartConsumer(reader *kafka.Reader, handler func([]byte), ctx context.Context) {
	defer kafkaConsumersWG.Done()

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
