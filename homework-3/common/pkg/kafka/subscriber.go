package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
)

type Subscriber struct {
	MessageHandler MessageHandlerFn
}

type MessageHandlerFn func(ctx context.Context, message *sarama.ConsumerMessage) error

func (s *Subscriber) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Subscriber) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (s *Subscriber) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.TODO()
	for msg := range claim.Messages() {
		err := s.MessageHandler(ctx, msg)
		if err != nil {
			logger.Error("failed to handle kafka message: " + err.Error())
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func StartSubscriber(brokers []string, groupId string, topics []string, msgHandler MessageHandlerFn, ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Return.Errors = true

	subscriber := Subscriber{MessageHandler: msgHandler}

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		return err
	}

	go func() {
		for err := range consumerGroup.Errors() {
			logger.Error(err.Error())
		}
	}()

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, topics, &subscriber); err != nil {
				logger.Error("failed to join and consume kafka topic: " + err.Error())
			}
			if ctx.Err() != nil {
				logger.Error(err.Error())
			}
		}
	}()
	return nil
}
