package proc

import (
	"context"
	"sync"
)

type Client struct {
	cli ConnectionAdapterClient
	buf []byte
	mut sync.Mutex
}

func NewClient(cli ConnectionAdapterClient) *Client {
	return &Client{
		cli: cli,
		buf: make([]byte, 0, 0xff),
	}
}

func (c *Client) OnReceivedBytes(ctx context.Context, msg []byte) {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, b := range msg {
		if b == 10 {
			c.parseBytes()

			if cap(c.buf) > 0xff {
				c.buf = make([]byte, 0, 0xff)
			} else {
				c.buf = c.buf[:0]
			}
		}

		c.buf = append(c.buf, b)
	}
}

func (c *Client) parseBytes() {
}

func (c *Client) OnTimerTimeout(ctx context.Context, ttp TimerType) {
	c.mut.Lock()
	defer c.mut.Unlock()
}

func (c *Client) OnReceivedMessage(ctx context.Context, msg *Message) {
	c.mut.Lock()
	defer c.mut.Unlock()
}
