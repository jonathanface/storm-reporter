package main

import (
	"context"

	"github.com/IBM/sarama"
)

type MockConsumerGroupSession struct{}

func (m *MockConsumerGroupSession) Claims() map[string][]int32 {
	return map[string][]int32{}
}

func (m *MockConsumerGroupSession) MemberID() string {
	return "mock-member"
}

func (m *MockConsumerGroupSession) GenerationID() int32 {
	return 1
}

func (m *MockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
}

func (m *MockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
}

func (m *MockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {}

func (m *MockConsumerGroupSession) Context() context.Context {
	return context.Background()
}

func (m *MockConsumerGroupSession) Commit() {}

type ConsumerGroupClaim interface {
	Topic() string
	Partition() int32
	InitialOffset() int64
	Messages() <-chan *sarama.ConsumerMessage
}
type MockConsumerGroupClaim struct {
	MessagesChannel chan *sarama.ConsumerMessage
}

func (m *MockConsumerGroupClaim) Topic() string {
	return "mock-topic"
}

func (m *MockConsumerGroupClaim) Partition() int32 {
	return 0
}

func (m *MockConsumerGroupClaim) InitialOffset() int64 {
	return sarama.OffsetOldest
}

func (m *MockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return m.MessagesChannel
}

func (m *MockConsumerGroupClaim) HighWaterMarkOffset() int64 {
	// Return a mock high watermark offset
	return 1000
}
