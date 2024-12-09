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

func main() {
	// Validate required environment variables
	if kafkaBrokers == "" || topic == "" || mongoURI == "" || mongoDBName == "" || mongoColl == "" {
		log.Fatal("KAFKA_BROKERS, PROCESSED_TOPIC, MONGO_URI, MONGO_DB, and MONGO_COLL must be set")
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

	http.HandleFunc("/messages", getMessagesHandler)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("API server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

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

		_, err := messagesColl.InsertOne(context.TODO(), message)
		if err != nil {
			log.Printf("Error inserting message into MongoDB: %v", err)
		} else {
			fmt.Printf("Message written to MongoDB: %v\n", message)
		}
	}
}

type StormType string

const (
	TORNADO StormType = "tornado"
	HAIL    StormType = "hail"
	WIND    StormType = "wind"
)

type StormReport struct {
	Time     int32     `json:"time" bson:"time"`
	Size     int32     `json:"size" bson:"size"`
	F_Scale  string    `json:"fScale" bson:"fScale"`
	Speed    int32     `json:"speed" bson:"speed"`
	Location string    `json:"location" bson:"location"`
	County   string    `json:"county" bson:"county"`
	State    string    `json:"state" bson:"state"`
	Lat      float64   `json:"lat" bson:"lat"`
	Lon      float64   `json:"lon" bson:"lon"`
	Comments string    `json:"comments" bson:"comments"`
	Type     StormType `json:"type" bson:"type"`
}

func getAllStormReports() ([]StormReport, error) {
	var reports []StormReport

	// Find all documents in the collection
	cursor, err := messagesColl.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve storm reports: %w", err)
	}
	defer cursor.Close(context.TODO())

	// Decode each document into a StormReport struct
	if err := cursor.All(context.TODO(), &reports); err != nil {
		return nil, fmt.Errorf("failed to decode storm reports: %w", err)
	}
	return reports, nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func returnError(w http.ResponseWriter, statusCode int, errorStr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error: errorStr,
	}
	json.NewEncoder(w).Encode(errorResponse)
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	reports, err := getAllStormReports()
	if err != nil {
		returnError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(reports) == 0 {
		returnError(w, http.StatusNotFound, "No storm reports found")
		return
	}

	jsonData, err := json.Marshal(reports)
	if err != nil {
		returnError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Write(jsonData)
}
