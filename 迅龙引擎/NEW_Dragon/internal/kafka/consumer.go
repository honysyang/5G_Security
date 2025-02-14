package kafka

import (
	"github.com/IBM/sarama"
)

// Consumer 接口定义了如何从 Kafka 消费消息
/*
定义了一个 Consumer 接口，它包含两个方法：Consume 和 Close。Consume 方法返回两个通道，一个用于接收 Kafka 消息，另一个用于接收错误信息。Close 方法用于关闭消费者连接。
*/
type Consumer interface {
	Consume() (<-chan *sarama.ConsumerMessage, <-chan error)
	Close() error
}

// KafkaConsumer 实现 Consumer 接口
/*
定义了一个 KafkaConsumer 结构体，它包含一个 sarama.Consumer 对象、一个主题（Topic）和分区（Partition）信息。
*/
type KafkaConsumer struct {
	Consumer  sarama.Consumer
	Topic     string
	Partition int32
}

// NewKafkaConsumer 创建一个新的 KafkaConsumer 实例
/*
定义了一个 NewKafkaConsumer 函数，用于创建 KafkaConsumer 的实例。它接收 Kafka 地址、主题和分区作为参数，并返回一个 Consumer 接口和错误。
*/
func NewKafkaConsumer(kafkaAddr string, topic string, partition int32) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	config.Consumer.IsolationLevel = sarama.ReadCommitted
	config.Version = sarama.V2_6_0_0 // 确保配置 Kafka 服务器兼容的版本
	consumer, err := sarama.NewConsumer([]string{kafkaAddr}, config)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumer{
		Consumer:  consumer,
		Topic:     topic,
		Partition: partition,
	}, nil
}

// Consume 实现了 Consumer 接口的方法
/*
Consume 方法实现了 Consumer 接口中的 Consume 方法，它返回两个通道，一个用于接收消息，一个用于接收错误。
*/
func (kc *KafkaConsumer) Consume() (<-chan *sarama.ConsumerMessage, <-chan error) {
	partitionConsumer, err := kc.Consumer.ConsumePartition(kc.Topic, kc.Partition, sarama.OffsetNewest)
	if err != nil {
		// 创建一个错误通道，发送错误，然后关闭通道
		errChan := make(chan error, 1)
		errChan <- err
		close(errChan)
		return nil, errChan
	}

	errChan := make(chan error)

	// 启动一个 goroutine 来监听 sarama.ConsumerError 通道，并将错误转换为 error 类型
	go func() {
		for err := range partitionConsumer.Errors() {
			errChan <- err
		}
		close(errChan)
	}()

	return partitionConsumer.Messages(), errChan
}

// Close 实现了 Consumer 接口的方法
func (kc *KafkaConsumer) Close() error {
	return kc.Consumer.Close()
}
