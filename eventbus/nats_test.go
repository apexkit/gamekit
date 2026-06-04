package eventbus

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const natsTestURL = "nats://localhost:4222"

// readBySubscribe 通过 Subscribe 接收消息
func TestNats(t *testing.T) {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	// 2. 创建 JetStream 接口
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 3. 创建 Stream（消息流）
	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"orders.*"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Stream created/verified:", stream.CachedInfo().Config.Name)

	// 4. 发布消息
	for i := 1; i <= 10; i++ {
		_, err := js.Publish(ctx, "orders.created", []byte(fmt.Sprintf("Order %d", i)))
		if err != nil {
			fmt.Printf("Publish error: %v", err)
		} else {
			fmt.Printf("Published: Order %d\n", i)
		}
	}

	// 5. 创建消费者（Pull Consumer）
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		// 指定 Durable 名称后会持久化消费进度，服务重启也不丢失
		Durable:   "order-processor",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 6. 消费消息 - 拉取方式
	fmt.Println("\n--- Fetching messages ---")
	msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	for msg := range msgs.Messages() {
		fmt.Printf("Received: %s\n", string(msg.Data()))
		msg.Ack() // 确认消息
	}

	// 7. 消费消息 - 持续监听方式（更常用）
	fmt.Println("\n--- Continuous consumption ---")
	wg := sync.WaitGroup{}
	wg.Add(5) // 只等待处理5条消息便于演示

	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		defer wg.Done()
		fmt.Printf("Consumed: %s\n", string(msg.Data()))
		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}
	defer consumeCtx.Stop()

	// 再发布几条消息来触发回调
	for i := 11; i <= 15; i++ {
		js.Publish(ctx, "orders.created", []byte(fmt.Sprintf("Order %d", i)))
	}

	wg.Wait()
	fmt.Println("All messages processed")

	// 等待中断信号优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Shutting down...")
}
