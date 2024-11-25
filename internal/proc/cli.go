package proc

import (
	context "context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	buf []byte
	obs string
	cli ConnectionAdapterClient
	mut sync.Mutex

	Conn string
	Log  *zap.Logger
}

func NewClient(conn string, adapter ConnectionAdapterClient, log *zap.Logger) *Client {
	return &Client{
		buf: make([]byte, 0, 0xff),
		cli: adapter,

		Conn: conn,
		Log:  log,
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buf = append(c.buf, b)
			continue
		}

		if err := c.exec(ctx); err != nil {
			return fmt.Errorf("exec: %v", err)
		}

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}

	return nil
}

func (c *Client) exec(ctx context.Context) error {
	if c.obs != "" {
		return nil
	}

	var cmd Command

	if err := cmd.Decode(c.buf); err != nil {
		return fmt.Errorf("decode: %v", err)
	}

	switch cmd.Method {
	case PUB:
		return c.publish(ctx, cmd.Topic, &cmd.Record)
	case SUB:
		return c.subscribe(ctx, cmd.Topic)
	case USB:
		return c.unsubscribe(ctx, cmd.Topic)
	case GET:
		return c.get(ctx, cmd.Topic)
	default:
		panic("impossible condition")
	}
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp TimerType) error { return nil }

func (c *Client) OnReceivedMessage(ctx context.Context, msg *Message) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.obs != "" {
		if c.obs != msg.GetTopic() {
			return nil
		}

		c.obs = ""

		if err := c.unsubscribe(ctx, msg.GetTopic()); err != nil {
			return fmt.Errorf("unsubscribe: %v", err)
		}
	}

	cmd, err := c.parseMessage(msg)

	if err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	if err := c.send(ctx, cmd); err != nil {
		return fmt.Errorf("send: %v", err)
	}

	return nil
}

func (c *Client) parseMessage(msg *Message) (*Command, error) {
	cmd := &Command{
		Topic: msg.GetTopic(),
	}

	if err := json.Unmarshal(msg.GetPayload(), &cmd.Record); err != nil {
		return nil, fmt.Errorf("payload: %v", err)
	}

	return cmd, nil
}

func (c *Client) send(ctx context.Context, cmd *Command) error {
	res, err := c.cli.Send(ctx, &SendBytesRequest{
		Conn:  c.Conn,
		Bytes: cmd.Encode(),
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) publish(ctx context.Context, top string, rec *Record) error {
	pay, err := json.Marshal(rec)

	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	res, err := c.cli.Publish(ctx, &PublishRequest{
		Conn:    c.Conn,
		Topic:   top,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, top string) error {
	res, err := c.cli.Subscribe(ctx, &SubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
		Qos:   2,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, top string) error {
	res, err := c.cli.Unsubscribe(ctx, &UnsubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
	})

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) get(ctx context.Context, top string) error {
	err := c.subscribe(ctx, top)

	if err != nil {
		return fmt.Errorf("subscribe: %v", err)
	}

	time.AfterFunc(5*time.Second, func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		if c.obs != "" {
			c.obs = ""

			c.unsubscribe(ctx, top)
			c.send(ctx, &Command{
				Topic: top,
				Record: Record{
					Timestamp: Time{time.Now()},
				},
			})
		}
	})

	return nil
}
