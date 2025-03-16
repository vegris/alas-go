package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type App struct {
	HTTPRoutes map[string]http.HandlerFunc
}

var Redis *redis.Client
var Kafka *kafka.Writer
var httpServer *http.Server

const EventsTopic = "kiwi-events"
const KeepAliveTopic = "keep-alive"

func Start(app *App) {
	startRedis()
	startKafka()
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

func startKafka() {
	const kafkaAddr = "localhost:9092"

	conn, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		log.Fatalf("Failed to establish Kafka connection: %v", err)
	}
	defer conn.Close()

    // Create needed topics
    topicsToCreate := [...]string{EventsTopic, KeepAliveTopic}
    topicConfigs := make([]kafka.TopicConfig, 0, len(topicsToCreate))

    for _, topicName := range topicsToCreate {
        topicConfigs = append(topicConfigs, kafka.TopicConfig{
            Topic: topicName,
            NumPartitions: 1,
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
    
	Kafka = &kafka.Writer{Addr: kafka.TCP(kafkaAddr), Async: true}
}

func shutdownKafka() {
	if err := Kafka.Close(); err == nil {
		log.Println("Kafka writer successfully closed!")
	} else {
		log.Printf("Failed to close Kafka writer: %v", err)
	}
}
