package types

// NATSPublisher 定義 NATS 發布接口
type NATSPublisher interface {
	Publish(topic string, data []byte) error
	GetTopics() TopicFormatter
}
