package shipment

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/event"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/internal/topic"
	ship "gitlab.ozon.dev/MShulgin/homework-3/shipment/pkg/shipment"
)

func StartKafkaSubscription(shipmentService Service, kafkaConf config.KafkaConfig, ctx context.Context) error {
	msgHandler := KafkaMsgHandler(shipmentService)
	topics := []string{topic.CommitDelivery, topic.NewShipmentDelivery, topic.CancelDelivery}
	if err := kafka.StartSubscriber(kafkaConf.Brokers, kafkaConf.GroupId, topics, msgHandler, ctx); err != nil {
		return err
	}
	return nil
}

func KafkaMsgHandler(shipmentService Service) func(ctx context.Context, msg *sarama.ConsumerMessage) error {
	cancelDeliveryHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		var m event.DeliveryCancel
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			return err
		}
		if err := shipmentService.CancelDelivery(m.ShipmentId); err != nil {
			return err
		}
		return nil
	}
	deliveryShipmentHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		var m ship.Shipment
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			return err
		}
		if err := shipmentService.AcceptDelivery(m); err != nil {
			return err
		}
		return nil
	}
	commitShipmentHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		var m ship.Shipment
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			return err
		}
		if err := shipmentService.CommitDelivery(m); err != nil {
			return err
		}
		return nil
	}

	msgHandler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		switch msg.Topic {
		case topic.CommitDelivery:
			return commitShipmentHandler(ctx, msg)
		case topic.NewShipmentDelivery:
			return deliveryShipmentHandler(ctx, msg)
		case topic.CancelDelivery:
			return cancelDeliveryHandler(ctx, msg)
		default:
			return nil
		}
	}
	return msgHandler
}
