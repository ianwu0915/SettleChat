package nats

import (
	"log"

	"github.com/ianwu0915/SettleChat/internal/types"
)

// Publish publish message to NATS
type NATSPublisher struct {
	natsManager *NATSManager
	env         string
	topics      types.TopicFormatter
}

func NewPublisher(natsManager *NATSManager, env string, topics types.TopicFormatter) *NATSPublisher {
	log.Printf("Creating new publisher for environment: %s", env)
	p := &NATSPublisher{
		natsManager: natsManager,
		env:         env,
		topics:      topics,
	}
	log.Printf("Publisher created successfully with env: %s", env)
	return p
}

// Publish implements the types.NATSPublisher interface
func (p *NATSPublisher) Publish(topic string, data []byte) error {
	return p.natsManager.Publish(topic, data)
}

// GetTopics returns the topic formatter
func (p *NATSPublisher) GetTopics() types.TopicFormatter {
	return p.topics
}
