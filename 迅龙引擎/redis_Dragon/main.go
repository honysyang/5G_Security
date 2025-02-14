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
		Addr:     "10.1.200.14:6379", // Redis 服务器地址和端口
		Password: "",                 // Redis 访问密码，如果没有可以为空字符串
		DB:       0,                  // 使用的 Redis 数据库编号，默认为 0
	})

	// 使用 Ping() 方法测试是否成功连接到 Redis 服务器
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		return
	}
	fmt.Println("Connected to Redis:", pong)

	// 订阅频道
	pubsub := rdb.Subscribe(ctx, "packet_channel")
	defer pubsub.Close()

	// 等待消息
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			time.Sleep(1 * time.Second) // 等待一段时间后重试
			continue
		}
		fmt.Printf("Received message from channel '%s': %02x\n", msg.Channel, msg.Payload)
	}

}
