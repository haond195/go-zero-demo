package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/service"
)

func consumeHandle(ctx context.Context, key, val string) error {
	fmt.Printf("\n======================================================\n")
	fmt.Printf("--> [KAFKA CONSUMER DA NHAN EVENT DON HANG]:\n")
	fmt.Printf("    Payload: %s\n", val)
	fmt.Printf("--> [EMAIL WORKER]: Dang gui email hoa don toi khach hang...\n")
	fmt.Printf("--> [THANH CONG]: Da hoan tat xu ly don hang tu Kafka!\n")
	fmt.Printf("======================================================\n\n")
	return nil
}

func main() {
	brokers := []string{"localhost:9092"}
	if bEnv := os.Getenv("KAFKA_BROKERS"); bEnv != "" {
		brokers = strings.Split(bEnv, ",")
	}

	qConf := kq.KqConf{
		Brokers:    brokers,
		Group:      "email-notification-group",
		Topic:      "order-created-topic",
		Offset:     "first",
		Consumers:  1,
		Processors: 1,
	}

	q := kq.MustNewQueue(qConf, kq.WithHandle(consumeHandle))
	defer q.Stop()

	services := service.NewServiceGroup()
	services.Add(q)
	fmt.Printf("--> [KAFKA CONSUMER WORKER] Dang lang nghe topic 'order-created-topic' tai %v...\n", brokers)
	services.Start()
}