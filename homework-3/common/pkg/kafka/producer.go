package kafka

import (
	"encoding/json"
	"github.com/Shopify/sarama"
)

func NewSyncProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	return producer, err
}

func SendMessage(producer sarama.SyncProducer, topic string, msg []byte) error {
	kafkaMsg := sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.StringEncoder(msg),
	}
	_, _, err := producer.SendMessage(&kafkaMsg)
	return err
}

func SendJsonMessage(producer sarama.SyncProducer, topic string, msg any) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return SendMessage(producer, topic, msgBytes)
}
