package services

import (
    "context"
    "encoding/json"
    
    "github.com/IBM/sarama"
    "warehouse/pkg/logger"
)

type EventService struct {
    producer sarama.SyncProducer
    topic    string
}

func NewEventService(brokers []string, topic string) (*EventService, error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Retry.Max = 5
    config.Producer.Return.Successes = true
    
    producer, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, err
    }
    
    return &EventService{
        producer: producer,
        topic:    topic,
    }, nil
}

func (s *EventService) PublishEvent(ctx context.Context, eventType string, event interface{}) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    msg := &sarama.ProducerMessage{
        Topic: s.topic,
        Key:   sarama.StringEncoder(eventType),
        Value: sarama.ByteEncoder(data),
    }
    
    _, _, err = s.producer.SendMessage(msg)
    if err != nil {
        logger.Error("Failed to publish event", err)
        return err
    }
    
    return nil
}

func (s *EventService) Close() error {
    return s.producer.Close()
}