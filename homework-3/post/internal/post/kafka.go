package post

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	confReader "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/event"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/post/internal/topic"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
)

func StartKafkaSubscription(postService Service, conf confReader.KafkaConfig, ctx context.Context) error {
	msgHandler := KafkaTopicHandler(postService)
	topics := []string{topic.CancelDelivery, topic.NewPostDelivery}
	if err := kafka.StartSubscriber(conf.Brokers, conf.GroupId, topics, msgHandler, ctx); err != nil {
		return err
	}
	return nil
}

func KafkaTopicHandler(postService Service) func(ctx context.Context, msg *sarama.ConsumerMessage) error {
	cancelDeliveryHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		var s event.DeliveryCancel
		if err := json.Unmarshal(msg.Value, &s); err != nil {
			return err
		}
		if err := postService.DeleteDelivery(s.ShipmentId); err != nil {
			return err
		}
		return nil
	}

	deliveryShipmentHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		var s shipment.Shipment
		if err := json.Unmarshal(msg.Value, &s); err != nil {
			return err
		}
		if err := postService.RegisterDelivery(s); err != nil {
			return err
		}
		return nil
	}

	msgHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		switch msg.Topic {
		case topic.CancelDelivery:
			return cancelDeliveryHandler(ctx, msg)
		case topic.NewPostDelivery:
			return deliveryShipmentHandler(ctx, msg)
		default:
			return nil
		}
	}
	return msgHandler
}
