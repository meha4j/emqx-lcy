package proc

import (
	context "context"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
)

type Client struct {
	conn string
	buff []byte
	mutx sync.Mutex

	adapter ConnectionAdapterClient
	log     *zap.Logger
}

func NewClient(conn string, adapter ConnectionAdapterClient, log *zap.Logger) *Client {
	return &Client{
		conn: conn,
		buff: make([]byte, 0, 0xff),

		adapter: adapter,
		log:     log,
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

		err := c.execute(ctx)

		if err != nil {
			c.log.Error("An error occurred while executing command.", zap.Error(err))
		}

		if cap(c.buff) > 0xff {
			c.buff = make([]byte, 0, 0xff)
		} else {
			c.buff = c.buff[:0]
		}
	}
}

func (c *Client) execute(ctx context.Context) error {
	var cmd Command

	err := cmd.Decode(c.buff)

	if err != nil {
		return err
	}

	switch cmd.Method {
	case PUB:
		return c.publish(ctx, &cmd)
	case SUB:
		return c.subscribe(ctx, &cmd)
	case USB:
		return c.unsubscribe(ctx, &cmd)
	case GET:
		return c.get(ctx, &cmd)
	default:
		panic("Must be already aborted.")
	}
}

func (c *Client) publish(ctx context.Context, cmd *Command) error {
	pay, err := json.Marshal(&cmd.Record)

	if err != nil {
		return err
	}

	_, err = c.adapter.Publish(ctx, &PublishRequest{
		Conn:    c.conn,
		Topic:   cmd.Topic,
		Qos:     0,
		Payload: pay,
	})

	return err
}

func (c *Client) subscribe(ctx context.Context, cmd *Command) error {
	_, err := c.adapter.Subscribe(ctx, &SubscribeRequest{
		Conn:  c.conn,
		Topic: cmd.Topic,
		Qos:   2,
	})

	return err
}

func (c *Client) unsubscribe(ctx context.Context, cmd *Command) error {
	_, err := c.adapter.Unsubscribe(ctx, &UnsubscribeRequest{
		Conn:  c.conn,
		Topic: cmd.Topic,
	})

	return err
}

func (c *Client) get(ctx context.Context, cmd *Command) error {
	err := c.subscribe(ctx, cmd)

	if err != nil {
		return err
	}

	return c.unsubscribe(ctx, cmd)
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp TimerType) {}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *Message) {
	c.mutx.Lock()
	defer c.mutx.Unlock()

	cmd, err := c.encodeMessage(msg)

	if err != nil {
		c.log.Error("An error occurred while encoding message.", zap.Error(err))
	}

	_, err = c.adapter.Send(ctx, &SendBytesRequest{
		Conn:  c.conn,
		Bytes: cmd,
	})

	if err != nil {
		c.log.Error("An error occurred while sending command.", zap.Error(err))
	}
}

func (c *Client) encodeMessage(msg *Message) ([]byte, error) {
	cmd := Command{
		Topic: msg.GetTopic(),
	}

	err := json.Unmarshal(msg.GetPayload(), &cmd.Record)

	if err != nil {
		return nil, err
	}

	return cmd.Encode()
}
