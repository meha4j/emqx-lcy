package emqx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	Bind string `json:"bind"`
}

type Handler struct {
	Addr string `json:"address"`
}

type GateUpdateRequest struct {
	Name       string     `json:"name"`
	Timeout    string     `json:"idle_timeout"`
	Mountpoint string     `json:"mountpoint"`
	Enable     bool       `json:"enable"`
	Statistics bool       `json:"enable_stats"`
	Server     Server     `json:"server"`
	Handler    Handler    `json:"handler"`
}

type HookUpdateRequest struct {
	Name      string `json:"name"`
	Enable    bool   `json:"enable"`
	Addr      string `json:"url"`
	Timeout   string `json:"request_timeout,omitempty"`
	Action    string `json:"failed_action,omitempty"`
	Reconnect string `json:"auto_reconnect,omitempty"`
	PoolSize  int    `json:"pool_size,omitempty"`
}

type options struct {
	addr *string
	user *string
	pass *string
	tout *time.Duration
	rmax *int
}

type Option func(*options) error

func WithUser(user string) Option {
	return func(o *options) error {
		o.user = &user
		return nil
	}
}

func WithPass(pass string) Option {
	return func(o *options) error {
		o.pass = &pass
		return nil
	}
}

func WithTimeout(tout string) Option {
	return func(o *options) error {
		res, err := time.ParseDuration(tout)

		if err != nil {
			return fmt.Errorf("parse: %v", err)
		}

		o.tout = &res
		return nil
	}
}

func WithRetries(rmax int) Option {
	return func(o *options) error {
		if rmax < 0 {
			return fmt.Errorf("negative retries count")
		}

		o.rmax = &rmax
		return nil
	}
}

type Client struct {
	Base string

	conn *http.Client
	user string
	pass string
	tout time.Duration
	rmax int
}

func NewClient(base string, opts ...Option) (*Client, error) {
	var opt options

	for _, exe := range opts {
		if err := exe(&opt); err != nil {
			return nil, fmt.Errorf("opt: %v", err)
		}
	}

	cli := &Client{
		Base: base,
		conn: &http.Client{},
	}

	if opt.user != nil {
		cli.user = *opt.user
	}

	if opt.pass != nil {
		cli.pass = *opt.pass
	}

	if opt.tout != nil {
		cli.tout = *opt.tout
	} else {
		cli.tout = 15 * time.Second
	}

	if opt.rmax != nil {
		cli.rmax = *opt.rmax
	} else {
		cli.rmax = 5
	}

	return cli, nil
}

func (c *Client) do(req *http.Request) (res *http.Response, err error) {
	if c.user != "" {
		req.SetBasicAuth(c.user, c.pass)
	}

	for r := 1; r <= c.rmax+1; r++ {
		res, err = c.conn.Do(req)

		if err == nil {
			break
		}

		slog.Error("req", "att", r, "rmax", c.rmax+1, "err", err)
		time.Sleep(c.tout)
	}

	return
}

func (c *Client) UpdateGate(r *GateUpdateRequest) error {
	pay, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("pay: %v", err)
	}

	url := c.Base + "/gateways/exproto"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(pay))

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.do(req)

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 204 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("res: %v", err)
		}

    slog.Debug("req update failed", "req", string(pay))

		return fmt.Errorf("%v", buf.String())
	}

	return nil
}

func (c *Client) UpdateHook(r *HookUpdateRequest) error {
	ok, err := c.checkHook(r.Name)

	if err != nil {
		return fmt.Errorf("check: %v", err)
	}

	pay, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("pay: %v", err)
	}

	url := c.Base + "/exhooks"
	mod := http.MethodPost

	if ok {
		url += "/extd"
		mod = http.MethodPut
	}

	req, err := http.NewRequest(mod, url, bytes.NewReader(pay))

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.do(req)

	if err != nil {
		return fmt.Errorf("req: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		var buf bytes.Buffer

		_, err := buf.ReadFrom(res.Body)

		if err != nil {
			return fmt.Errorf("res: %v", err)
		}

		return fmt.Errorf("%v", buf.String())
	}

	return nil
}

func (c *Client) checkHook(name string) (bool, error) {
	url := c.Base + "/exhooks/" + name
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return false, fmt.Errorf("req: %v", err)
	}

	res, err := c.do(req)

	if err != nil {
		return false, fmt.Errorf("req: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}
