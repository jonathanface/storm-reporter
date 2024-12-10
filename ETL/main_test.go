package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
)

func TestTransformData_ValidData(t *testing.T) {
	rawJSON := `{
		"date": "1733773195",
		"Time": "1200",
		"Size": "3",
		"F_Scale": "",
		"Speed": "0",
		"Location": "Boston",
		"County": "some county",
		"State": "MA",
		"Lat": "43.227946016550874",
		"Lon": "-78.49762842469941",
		"Comments": "it's really bad",
		"Type": "hail"
	}`

	_, err := transformData(rawJSON)
	assert.NoError(t, err, "transformData should not return an error for valid input")

	var result StormReport
	err = json.Unmarshal([]byte(rawJSON), &result)
	assert.NoError(t, err, "Unmarshaling transformed data should not return an error")

	assert.Equal(t, "1733773195", result.Date)
	assert.Equal(t, int32(1200), result.Time)
	assert.Equal(t, float64(3), result.Size)
	assert.Equal(t, "hail", string(result.Type))
	assert.Equal(t, "Boston", result.Location)
	assert.Equal(t, "it's really bad", result.Comments)
}

func TestTransformData_InvalidHeader(t *testing.T) {
	rawJSON := `{
		"Time": "time",
		"Size": "3",
		"F_Scale": "",
		"Speed": "0",
		"Location": "Boston",
		"County": "some county",
		"State": "MA",
		"Lat": "43.227946016550874",
		"Lon": "-78.49762842469941",
		"Comments": "it's really bad",
		"Type": "hail"
	}`

	transformed, err := transformData(rawJSON)
	assert.Error(t, err, "transformData should return an error for invalid header field")
	assert.Empty(t, transformed, "Transformed data should be empty for invalid input")
}

func TestKafkaProducer_SuccessfulSend(t *testing.T) {
	producer := mocks.NewSyncProducer(t, nil)
	defer producer.Close()

	producer.ExpectSendMessageAndSucceed()
	expectedPartition := int32(0)
	expectedOffset := int64(1)

	message := `{"date": "1733773195", "time": "1200", "type": "hail"}`
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "processed-topic",
		Value: sarama.StringEncoder(message),
	})
	fmt.Println(partition, offset)

	assert.NoError(t, err, "Producer should not return an error on successful send")
	assert.Equal(t, expectedPartition, partition, "Partition should match the expected value")
	assert.Equal(t, expectedOffset, offset, "Offset should match the expected value")
}

func TestKafkaProducer_FailedSend(t *testing.T) {
	producer := mocks.NewSyncProducer(t, nil)
	defer producer.Close()

	producer.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)

	message := `{"date": "1733773195", "time": "1200", "type": "hail"}`
	_, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "processed-topic",
		Value: sarama.StringEncoder(message),
	})

	assert.Error(t, err, "Producer should return an error when sending fails")
	assert.Equal(t, sarama.ErrOutOfBrokers, err, "Error should be 'OutOfBrokers'")
}

func TestETLHandler_ConsumeClaim(t *testing.T) {
	// Mock producer
	mockProducer := mocks.NewSyncProducer(t, nil)
	defer mockProducer.Close()

	// Mock consumer group session
	session := &MockConsumerGroupSession{}

	// Mock consumer group claim
	mockClaim := &MockConsumerGroupClaim{
		MessagesChannel: make(chan *sarama.ConsumerMessage, 1),
	}

	rawJSON := `{"date":"2023-12-09","time":"1200","type":"tornado","location":"Test City","fScale":"F5","size":"0.2","speed":"5","Lat":"43.227946016550874","Lon":"-78.49762842469941"}`

	// Add a test message
	mockClaim.MessagesChannel <- &sarama.ConsumerMessage{
		Value: []byte(rawJSON),
	}
	close(mockClaim.MessagesChannel)

	// Create ETLHandler
	handler := &ETLHandler{
		producer:       mockProducer,
		rawTopic:       "test-raw-topic",
		processedTopic: "test-processed-topic",
	}

	// Verify producer sent message
	expected := `{"comments":"","state":"","county":"","date":"2023-12-09","time":1200,"type":"tornado","location":"Test City","fScale":"F5","size":0.2,"speed":5,"lat":43.227946016550874,"lon":-78.49762842469941}`
	mockProducer.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		assert.JSONEq(t, expected, string(val))
		return nil
	})

	// Test ConsumeClaim
	err := handler.ConsumeClaim(session, mockClaim)
	assert.NoError(t, err, "ConsumeClaim should not return an error")

}
