// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "10.1.200.14:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	fmt.Println("Start!")

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	} else {
		fmt.Println("Connected to Redis successfully!")
	}

	// 订阅Redis频道
	pubsub := rdb.Subscribe(ctx, "packet_channel")
	defer func() {
		if err := pubsub.Close(); err != nil {
			log.Printf("Failed to close pubsub: %v", err)
		}
	}()

	// 检查是否有错误
	_, err := pubsub.Receive(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	fmt.Println("Subscribed to packet_channel")

	// 消费数据
	ch := pubsub.Channel()
	for msg := range ch {
		fmt.Printf("Received message: %s\n", msg.Payload)
		// 处理接收到的数据包
	}

}

