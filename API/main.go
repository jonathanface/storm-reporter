package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/jonathanface/storm-reporter/API/dao"
	"github.com/jonathanface/storm-reporter/API/middleware"
	"github.com/jonathanface/storm-reporter/API/routes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	kafkaBrokers = os.Getenv("KAFKA_BROKERS")
	topic        = os.Getenv("PROCESSED_TOPIC")
	mongoURI     = os.Getenv("MONGO_URI")
	mongoDBName  = os.Getenv("MONGO_DB")
	mongoColl    = os.Getenv("MONGO_COLL")
	client       *mongo.Client
	messagesColl *mongo.Collection
)

func consumeFromKafka() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in consumeFromKafka: %v", r)
		}
	}()

	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	var consumer sarama.Consumer
	var err error

	// Retry connecting to Kafka
	for {
		consumer, err = sarama.NewConsumer([]string{kafkaBrokers}, config)
		if err != nil {
			log.Printf("Error creating Kafka consumer: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Error creating partition consumer: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
		go consumeFromKafka() // Retry by restarting the consumer
		return
	}
	defer partitionConsumer.Close()

	fmt.Println("Consuming messages from Kafka...")
	for msg := range partitionConsumer.Messages() {
		fmt.Printf("Received message: %s\n", string(msg.Value))

		// Store message in MongoDB
		var message bson.M
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		message["kafkaPartition"] = msg.Partition
		message["kafkaOffset"] = msg.Offset

		//attempting to prevent duplicate storms
		fmt.Println("write message:", message)
		filter := bson.M{"time": message["time"], "type": message["type"], "location": message["location"], "lat": message["lat"], "lon": message["lon"]}
		_, err := messagesColl.UpdateOne(
			context.TODO(),
			filter,
			bson.M{"$set": message},
			options.Update().SetUpsert(true),
		)
		//_, err := messagesColl.InsertOne(context.TODO(), message)
		if err != nil {
			log.Printf("Error inserting message into MongoDB: %v", err)
		} else {
			fmt.Printf("Message written to MongoDB: %v\n", message)
		}
	}
}

func main() {
	if kafkaBrokers == "" || mongoURI == "" || mongoDBName == "" || mongoColl == "" {
		log.Fatal("Environment variables KAFKA_BROKERS, MONGO_URI, MONGO_DB, and MONGO_COLL must be set")
	}

	// Initialize MongoDB
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	messagesColl = client.Database(mongoDBName).Collection(mongoColl)
	fmt.Println("Connected to MongoDB")

	// Start Kafka consumer in a goroutine
	go consumeFromKafka()

	// Initialize DAO
	dao, err := dao.NewStormDAO(mongoURI, mongoDBName, mongoColl)
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}
	defer dao.Disconnect()

	// routes with middleware
	mux := http.NewServeMux()
	middlewareContext := middleware.WithDAOContext(dao)
	mux.Handle("/messages", middlewareContext(routes.GetMessagesHandler))

	// Start
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
