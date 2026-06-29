package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

const natsStreamName = "eventbus"

type NatsBus struct {
	url    string
	conn   *nats.Conn
	js     nats.JetStreamContext
	connMu sync.Mutex
	subMu  sync.Mutex
	subs   map[string]*nats.Subscription
}

// NewNatsBus 创建一个 NATS 事件总线（首次 Publish/Subscribe 时连接）。
func NewNatsBus(url string) EventBus {
	return &NatsBus{
		url:  url,
		subs: make(map[string]*nats.Subscription),
	}
}

// NewConnectedNatsBus 创建并立即连接 NATS，用于启动阶段验收。
func NewConnectedNatsBus(url string) (EventBus, error) {
	b := &NatsBus{
		url:  url,
		subs: make(map[string]*nats.Subscription),
	}
	if err := b.connect(); err != nil {
		return nil, fmt.Errorf("nats connect %s: %w", url, err)
	}
	return b, nil
}

func (b *NatsBus) connect() error {
	b.connMu.Lock()
	defer b.connMu.Unlock()

	if b.conn != nil && b.conn.IsConnected() {
		return nil
	}
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
		b.js = nil
	}

	conn, err := nats.Connect(b.url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				log.Errorf("NATS disconnected from %s: %v", b.url, err)
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Infof("NATS reconnected to %s", b.url)
		}),
	)
	if err != nil {
		return err
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return err
	}

	b.conn = conn
	b.js = js
	log.Infof("Connected to NATS %s", b.url)
	return nil
}

func (b *NatsBus) ensureStream(subject string) error {
	if err := b.connect(); err != nil {
		return err
	}

	info, err := b.js.StreamInfo(natsStreamName)
	if err != nil {
		_, err = b.js.AddStream(&nats.StreamConfig{
			Name:      natsStreamName,
			Subjects:  []string{subject},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
		})
		if err != nil {
			return err
		}
		log.Infof("Created NATS stream %s for subject %s", natsStreamName, subject)
		return nil
	}

	for _, s := range info.Config.Subjects {
		if subject == s {
			return nil
		}
	}

	subjects := append(info.Config.Subjects, subject)
	cfg := info.Config
	cfg.Subjects = subjects
	_, err = b.js.UpdateStream(&cfg)
	if err != nil {
		return err
	}
	log.Infof("Updated NATS stream %s with subject %s", natsStreamName, subject)
	return nil
}

// Publish 发布事件
func (b *NatsBus) Publish(ctx context.Context, topic string, evt *Event) error {
	if evt == nil {
		return fmt.Errorf("topic %s publish event is nil", topic)
	}
	if err := b.ensureStream(topic); err != nil {
		return err
	}

	data, err := evt.Encode()
	if err != nil {
		return err
	}

	_, err = b.js.PublishMsg(&nats.Msg{
		Subject: topic,
		Data:    data,
		Header:  nats.Header{nats.MsgIdHdr: []string{evt.EventID}},
	}, nats.Context(ctx))
	if err != nil {
		log.Errorf("NATS publish failed for topic %s: %v", topic, err)
	}
	return err
}

// Subscribe 订阅事件
func (b *NatsBus) Subscribe(ctx context.Context, topic, group string, handler EventHandler) error {
	key := fmt.Sprintf("%s-%s", topic, group)

	b.subMu.Lock()
	if _, exists := b.subs[key]; exists {
		b.subMu.Unlock()
		return fmt.Errorf("already subscribed to topic %s with group %s", topic, group)
	}

	if err := b.ensureStream(topic); err != nil {
		b.subMu.Unlock()
		return err
	}

	durable := fmt.Sprintf("%s_%s", topic, group)
	sub, err := b.js.QueueSubscribe(topic, group, func(msg *nats.Msg) {
		evt, err := DecodeEvent(msg.Data)
		if err != nil {
			log.Infof("Decode event error (topic: %s, group: %s): %v", topic, group, err)
			_ = msg.Nak()
			return
		}

		if err := handler(ctx, evt); err != nil {
			log.Infof("Event handler error: %v (event: %s, topic: %s, group: %s)",
				err, evt.Type, topic, group)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	},
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
		nats.MaxDeliver(5),
	)
	if err != nil {
		b.subMu.Unlock()
		return err
	}

	b.subs[key] = sub
	log.Infof("Created new NATS subscriber for topic %s, group %s", topic, group)
	b.subMu.Unlock()

	go func() {
		<-ctx.Done()
		log.Infof("Stopping consumer for topic %s, group %s", topic, group)
		b.subMu.Lock()
		delete(b.subs, key)
		b.subMu.Unlock()
		_ = sub.Unsubscribe()
	}()

	return nil
}

func (b *NatsBus) Close() error {
	b.subMu.Lock()
	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}
	b.subs = make(map[string]*nats.Subscription)
	b.subMu.Unlock()

	b.connMu.Lock()
	defer b.connMu.Unlock()
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
		b.js = nil
	}
	return nil
}
