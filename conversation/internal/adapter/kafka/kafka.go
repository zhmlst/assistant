package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	pkgkafka "github.com/zhmlst/assistant/conversation/pkg/kafka"
	gokafka "github.com/zhmlst/assistant/go/kafka"
	"github.com/zhmlst/assistant/conversation/internal/domain"
	conversationv1 "github.com/zhmlst/assistant/conversation/pkg/conversation/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type eventPublisher struct {
	procuder *kafka.Producer
}

func New(producer *kafka.Producer) *eventPublisher {
	return &eventPublisher{
		procuder: producer,
	}
}

func (p *eventPublisher) MessageCreated(ctx context.Context, msg *domain.Message) error {
	role, err := conversationv1.RoleToProto(msg.Role)
	if err != nil {
		return fmt.Errorf("message role to proto: %w", err)
	}

	protomsg := conversationv1.Message{
		Id:             msg.ID[:],
		ParentId:       msg.ParentID[:],
		ConversationId: msg.ConversationID[:],
		Role:           role,
		Text:           msg.Text,
		CreateTime:     timestamppb.New(msg.CreatedAt),
	}

	kafkamsg := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     new(pkgkafka.TopicMessage),
			Partition: kafka.PartitionAny,
		},
		Key:     msg.ID[:],
		Headers: []kafka.Header{
			{Key: gokafka.EventTypeKey, Value: []byte(pkgkafka.EventMessageCreated)},
		},
	}

	kafkamsg.Value, err = proto.Marshal(&protomsg)
	if err != nil {
		return fmt.Errorf("marshal proto message: %w", err)
	}

	if err := p.procuder.Produce(&kafkamsg, nil); err != nil {
		return fmt.Errorf("produce kafka message: %w", err)
	}

	return nil
}
