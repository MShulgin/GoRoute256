package offer

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	confReader "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/offer/internal/topic"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
)

func StartKafkaSubscription(offerService Service, conf confReader.KafkaConfig, ctx context.Context) error {
	msgHandler := KafkaTopicHandlers(offerService)
	topics := []string{topic.RemoveReserve}
	return kafka.StartSubscriber(conf.Brokers, conf.GroupId, topics, msgHandler, ctx)
}

func KafkaTopicHandlers(offerService Service) func(ctx context.Context, msg *sarama.ConsumerMessage) error {
	removeReserveHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		var s shipment.Shipment
		if err := json.Unmarshal(msg.Value, &s); err != nil {
			return err
		}
		if err := offerService.RemoveReserve(s); err != nil {
			return err
		}
		return nil
	}
	msgHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		switch msg.Topic {
		case topic.RemoveReserve:
			return removeReserveHandler(ctx, msg)
		default:
			return nil
		}
	}
	return msgHandler
}
