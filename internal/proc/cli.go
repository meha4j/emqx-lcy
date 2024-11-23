package proc

import (
	context "context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	ERR = "Runtime error."
)

type Client struct {
	buf []byte
	obs string
	mut sync.Mutex

	Conn    string
	Adapter ConnectionAdapterClient
	Log     *zap.Logger
}

func NewClient(conn string, adapter ConnectionAdapterClient, log *zap.Logger) *Client {
	return &Client{
		buf: make([]byte, 0, 0xff),

		Conn:    conn,
		Adapter: adapter,
		Log:     log,
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, b := range msg {
		if b != 10 {
			c.buf = append(c.buf, b)
			continue
		}

		err := c.exec(ctx)

		if err != nil {
			c.Log.Error(ERR, zap.Error(err))
		}

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}
}

func (c *Client) exec(ctx context.Context) error {
	if c.obs != "" {
		return nil
	}

	var cmd Command

	err := cmd.Decode(c.buf)

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

	res, err := c.Adapter.Publish(ctx, &PublishRequest{
		Conn:    c.Conn,
		Topic:   cmd.Topic,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) subscribe(ctx context.Context, cmd *Command) error {
	res, err := c.Adapter.Subscribe(ctx, &SubscribeRequest{
		Conn:  c.Conn,
		Topic: cmd.Topic,
		Qos:   2,
	})

	if err != nil {
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, cmd *Command) error {
	res, err := c.Adapter.Unsubscribe(ctx, &UnsubscribeRequest{
		Conn:  c.Conn,
		Topic: cmd.Topic,
	})

	if err != nil {
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		return fmt.Errorf(res.GetMessage())
	}

	return nil
}

func (c *Client) get(ctx context.Context, cmd *Command) error {
	err := c.subscribe(ctx, cmd)

	if err != nil {
		return err
	}

	time.AfterFunc(5*time.Second, func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		if c.obs != "" {
			var empty Command

			empty.Topic = cmd.Topic
			empty.Timestamp.Now()

			enc, _ := empty.Encode()
			res, err := c.Adapter.Send(ctx, &SendBytesRequest{
				Conn:  c.Conn,
				Bytes: enc,
			})

			if err != nil {
				c.Log.Error(ERR, zap.Error(err))
			}

			if res.GetCode() != ResultCode_SUCCESS {
				c.Log.Error(ERR, zap.String("error", res.GetMessage()))
			}

			c.obs = ""
			err = c.unsubscribe(ctx, cmd)

			if err != nil {
				c.Log.Error(ERR, zap.Error(err))
			}
		}
	})

	return nil
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp TimerType) {}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *Message) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.obs != "" {
		if c.obs != msg.GetTopic() {
			return
		}

		c.obs = ""
	}

	cmd, err := c.encodeMessage(msg)

	if err != nil {
		c.Log.Error("An error occurred while encoding message.", zap.Error(err))
	}

	_, err = c.Adapter.Send(ctx, &SendBytesRequest{
		Conn:  c.Conn,
		Bytes: cmd,
	})

	if err != nil {
		c.Log.Error("An error occurred while sending command.", zap.Error(err))
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
