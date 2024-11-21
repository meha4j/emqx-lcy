package proc

import (
	context "context"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

type Client struct {
	log *zap.Logger
	cac ConnectionAdapterClient

	conn string
	buff []byte
	mutx sync.Mutex
}

func NewClient(conn string, cac ConnectionAdapterClient, log *zap.Logger) *Client {
	return &Client{
		log:  log,
		cac:  cac,
		conn: conn,
		buff: make([]byte, 0, 0xff),
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buff = append(c.buff, b)
			continue
		}

		err := c.processEvent(ctx)

		if err != nil {
			c.log.Error("An error occurred while processing event.", zap.Error(err))
		}

		if cap(c.buff) > 0xff {
			c.buff = make([]byte, 0, 0xff)
		} else {
			c.buff = c.buff[:0]
		}
	}
}

func (c *Client) processEvent(ctx context.Context) error {
	var event Event

	err := event.Decode(c.buff)

	if err != nil {
		return err
	}

	switch event.Method {
	case PUB:
		return c.processPublish(ctx, &event)
	case SUB:
		return c.processSubscribe(ctx, &event)
	case USB:
		return c.processUnsubscribe(ctx, &event)
	case GET:
		return c.processGet(ctx, &event)
	default:
		panic("Must be already aborted.")
	}
}

func (c *Client) processPublish(ctx context.Context, event *Event) error {
	pay, err := json.Marshal(event.Record)

	if err != nil {
		return err
	}

	_, err = c.cac.Publish(ctx, &PublishRequest{
		Conn:    c.conn,
		Topic:   event.Topic,
		Qos:     0,
		Payload: pay,
	})

	return err
}

func (c *Client) processSubscribe(ctx context.Context, event *Event) error {
	_, err := c.cac.Subscribe(ctx, &SubscribeRequest{
		Conn:  c.conn,
		Topic: event.Topic,
		Qos:   2,
	})

	return err
}

func (c *Client) processUnsubscribe(ctx context.Context, event *Event) error {
	_, err := c.cac.Unsubscribe(ctx, &UnsubscribeRequest{
		Conn:  c.conn,
		Topic: event.Topic,
	})

	return err
}

func (c *Client) processGet(ctx context.Context, event *Event) error {
	err := c.processSubscribe(ctx, event)

	if err != nil {
		return err
	}

	return c.processUnsubscribe(ctx, event)
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp TimerType) {}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *Message) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	event, err := c.encodeMessage(msg)

	if err != nil {
		c.log.Error("An error occurred while encoding EMQX message.", zap.Error(err))
	}

	_, err = c.cac.Send(ctx, &SendBytesRequest{
		Conn:  c.conn,
		Bytes: event,
	})

	if err != nil {
		c.log.Error("An error occurred while sending event.", zap.Error(err))
	}
}

func (c *Client) encodeMessage(msg *Message) ([]byte, error) {
	event := Event{
		Topic:  msg.GetTopic(),
		Method: PUB,
	}

	err := json.Unmarshal(msg.GetPayload(), &event.Record)

	if err != nil {
		return nil, err
	}

	return event.Encode()
}
