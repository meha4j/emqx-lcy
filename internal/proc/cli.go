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
	mut sync.Mutex
	cli ConnectionAdapterClient

	Conn string
	Log  *zap.Logger
}

func NewClient(conn string, adapter ConnectionAdapterClient, log *zap.Logger) *Client {
	log.Info("New connection.")

	return &Client{
		buf: make([]byte, 0, 0xff),
		cli: adapter,

		Conn: conn,
		Log:  log,
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

		c.exec(ctx)

		if cap(c.buf) > 0xff {
			c.buf = make([]byte, 0, 0xff)
		} else {
			c.buf = c.buf[:0]
		}
	}
}

func (c *Client) exec(ctx context.Context) error {
	if c.obs != "" {
		c.Log.Warn("Execution cancelled due to active GET request.")
		return nil
	}

	var cmd Command

	err := cmd.Decode(c.buf)

	if err != nil {
		c.Log.Error("Could not decode command.", zap.Error(err), zap.String("cmd", string(c.buf)))
		return err
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
		panic("Impossible condition.")
	}
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
		c.unsubscribe(ctx, msg.GetTopic())
	}

	cmd, _ := c.parseMessage(msg)
	enc, err := cmd.Encode()

	if err != nil {
		c.Log.Error("Could not encode message.", zap.Error(err))
		return
	}

	_, err = c.cli.Send(ctx, &SendBytesRequest{
		Conn:  c.Conn,
		Bytes: enc,
	})

	if err != nil {
		c.Log.Error("Could not send message.", zap.Error(err))
	}
}

func (c *Client) parseMessage(msg *Message) (*Command, error) {
	cmd := &Command{
		Topic: msg.GetTopic(),
	}

	err := json.Unmarshal(msg.GetPayload(), &cmd.Record)

	if err != nil {
		c.Log.Error("Could not parse message.", zap.Error(err))
		return nil, err
	}

	return cmd, nil
}

func (c *Client) send(ctx context.Context, cmd *Command) error {
	bytes, err := cmd.Encode()

	if err != nil {
		c.Log.Error("Could not encode command.", zap.Error(err))
		return err
	}

	res, err := c.cli.Send(ctx, &SendBytesRequest{
		Conn:  c.Conn,
		Bytes: bytes,
	})

	if err != nil {
		c.Log.Error("Could not send request.", zap.Error(err), zap.Any("cmd", cmd))
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		err := fmt.Errorf(res.GetMessage())
		c.Log.Error("Unsuccessful request.", zap.Error(err))
		return err
	}

	return nil
}

func (c *Client) publish(ctx context.Context, top string, rec *Record) error {
	pay, err := json.Marshal(rec)

	if err != nil {
		c.Log.Error("Could not encode record.", zap.Error(err), zap.Any("record", rec))
		return err
	}

	res, err := c.cli.Publish(ctx, &PublishRequest{
		Conn:    c.Conn,
		Topic:   top,
		Qos:     0,
		Payload: pay,
	})

	if err != nil {
		c.Log.Error("Could not send request.", zap.Error(err))
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		err := fmt.Errorf(res.GetMessage())
		c.Log.Error("Unsuccessful request.", zap.Error(err))
		return err
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
		c.Log.Error("Could not send request.", zap.Error(err))
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		err := fmt.Errorf(res.GetMessage())
		c.Log.Error("Unsuccessful request.", zap.Error(err))
		return err
	}

	return nil
}

func (c *Client) unsubscribe(ctx context.Context, top string) error {
	res, err := c.cli.Unsubscribe(ctx, &UnsubscribeRequest{
		Conn:  c.Conn,
		Topic: top,
	})

	if err != nil {
		c.Log.Error("Could not send request.", zap.Error(err))
		return err
	}

	if res.GetCode() != ResultCode_SUCCESS {
		err := fmt.Errorf(res.GetMessage())
		c.Log.Error("Unsuccessful request.", zap.Error(err))
		return err
	}

	return nil
}

func (c *Client) get(ctx context.Context, top string) error {
	err := c.subscribe(ctx, top)

	if err != nil {
		return err
	}

	time.AfterFunc(5*time.Second, func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		if c.obs != "" {
			c.unsubscribe(ctx, top)
			c.send(ctx, &Command{
				Topic: top,
				Record: Record{
					Timestamp: Now(),
				},
			})

			c.obs = ""
		}
	})

	return nil
}
