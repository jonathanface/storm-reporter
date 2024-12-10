package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

var (
	brokers        = os.Getenv("KAFKA_BROKERS")
	rawTopic       = os.Getenv("RAW_TOPIC")
	processedTopic = os.Getenv("PROCESSED_TOPIC")
)

func main() {
	if brokers == "" || rawTopic == "" || processedTopic == "" {
		log.Fatal("KAFKA_BROKERS, RAW_TOPIC, and PROCESSED_TOPIC environment variables must be set")
	}

	config := sarama.NewConfig()
	config.Producer.MaxMessageBytes = 209715200 // 200 MB
	config.Consumer.Fetch.Max = 209715200       // 200 MB
	config.Consumer.MaxWaitTime = 10 * time.Millisecond
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Producer.Return.Successes = true

	log.Println("Starting ETL service...")
	runETL(brokers, rawTopic, processedTopic, config)
}

func runETL(brokers, rawTopic, processedTopic string, config *sarama.Config) {
	consumerGroup, err := sarama.NewConsumerGroup([]string{brokers}, "etl-consumer-group", config)
	if err != nil {
		log.Fatalf("Error creating consumer group: %v", err)
	}
	defer consumerGroup.Close()

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()

	ctx := context.Background()
	handler := &ETLHandler{
		producer:       producer,
		rawTopic:       rawTopic,
		processedTopic: processedTopic,
	}

	log.Println("Listening for messages...")
	for {
		if err := consumerGroup.Consume(ctx, []string{rawTopic}, handler); err != nil {
			log.Printf("Error consuming messages: %v", err)
		}
	}
}

// ETLHandler implements sarama.ConsumerGroupHandler
type ETLHandler struct {
	producer       sarama.SyncProducer
	rawTopic       string
	processedTopic string
}

// Setup is called before consuming messages
func (h *ETLHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called after consuming messages
func (h *ETLHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from the Kafka topic
func (h *ETLHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Received message: %s", string(message.Value))
		transformedMessage, err := transformData(string(message.Value))
		if err != nil {
			log.Printf("Error transforming data: %v", err)
			continue
		}
		log.Printf("Transformed message: %s", transformedMessage)
		partition, offset, err := h.producer.SendMessage(&sarama.ProducerMessage{
			Topic: h.processedTopic,
			Value: sarama.StringEncoder(transformedMessage),
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
			continue
		}

		log.Printf("Message sent to topic %s: %s (partition=%d, offset=%d)", h.processedTopic, transformedMessage, partition, offset)
		session.MarkMessage(message, "")
	}
	return nil
}

type StormType string

const (
	TORNADO StormType = "tornado"
	HAIL    StormType = "hail"
	WIND    StormType = "wind"
)

type StormReport struct {
	Date     string    `json:"date"`
	Time     int32     `json:"time"`
	Size     int32     `json:"size"`
	F_Scale  string    `json:"fScale"`
	Speed    int32     `json:"speed"`
	Location string    `json:"location"`
	County   string    `json:"county"`
	State    string    `json:"state"`
	Lat      float64   `json:"lat"`
	Lon      float64   `json:"lon"`
	Comments string    `json:"comments"`
	Type     StormType `json:"type"`
}

func (sr *StormReport) UnmarshalJSON(data []byte) error {
	// Define a temporary struct matching the incoming JSON keys
	type tempStormReport struct {
		Date     string `json:"date"`
		Time     string `json:"Time"`
		Size     string `json:"Size"`
		F_Scale  string `json:"F_Scale"`
		Speed    string `json:"Speed"`
		Location string `json:"Location"`
		County   string `json:"County"`
		State    string `json:"State"`
		Lat      string `json:"Lat"`
		Lon      string `json:"Lon"`
		Comments string `json:"Comments"`
		Type     string `json:"Type"`
	}

	var temp tempStormReport
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if val, err := strconv.Atoi(temp.Time); err == nil {
		sr.Time = int32(val)
	} else {
		return err
	}
	if len(temp.Size) > 0 {
		if val, err := strconv.Atoi(temp.Size); err == nil {
			sr.Size = int32(val)
		} else {
			return err
		}
	}
	if len(temp.Speed) > 0 {
		if val, err := strconv.Atoi(temp.Speed); err == nil {
			sr.Speed = int32(val)
		} else {
			return err
		}
	}
	if val, err := strconv.ParseFloat(temp.Lat, 64); err == nil {
		sr.Lat = float64(val)
	} else {
		return err
	}
	if val, err := strconv.ParseFloat(temp.Lon, 64); err == nil {
		sr.Lon = float64(val)
	} else {
		return err
	}
	sr.Date = temp.Date
	sr.F_Scale = temp.F_Scale
	sr.Location = temp.Location
	sr.County = temp.County
	sr.State = temp.State
	sr.Comments = temp.Comments
	sr.Type = StormType(temp.Type)
	return nil
}

func transformData(data string) (string, error) {

	var raw map[string]interface{}
	err := json.Unmarshal([]byte(data), &raw)
	if err != nil {
		fmt.Printf("Error unmarshaling raw JSON: %v\n", err)
		return "", err
	}

	// Validate the "Time" field
	if timeValue, ok := raw["Time"].(string); ok && strings.ToLower(timeValue) == "time" {
		return "", fmt.Errorf("invalid data: header field detected")
	}

	// Now unmarshal into StormReport struct
	var report StormReport
	err = json.Unmarshal([]byte(data), &report)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON into StormReport: %v\n", err)
		return "", err
	}

	// Marshal the validated and transformed StormReport back to JSON
	transformedData, err := json.Marshal(report)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return "", err
	}

	return string(transformedData), nil
}
