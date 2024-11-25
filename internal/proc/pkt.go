package proc

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Stamp = "02.01.2006 15_04_05.000"

	PUB = iota + 1
	SUB
	USB
	GET
)

type Time struct {
	time.Time
}

func (t *Time) Parse(value string) error {
	tm, err := time.ParseInLocation(Stamp, value, time.Local)

	if err != nil {
		return fmt.Errorf("parse time: %v", err)
	}

	t.Time = tm

	return nil
}

func (t *Time) Format() string {
	return t.Time.Format(Stamp)
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

func (t *Time) UnmarshalJSON(json []byte) error {
	milli, err := strconv.ParseInt(string(json), 10, 64)

	if err != nil {
		return fmt.Errorf("unmarshal time: %v", err)
	}

	t.Time = time.UnixMilli(milli)

	return nil
}

type Record struct {
	Value     any  `json:"value,omitempty"`
	Timestamp Time `json:"timestamp"`
}

type Command struct {
	Topic  string
	Method int

	Record
}

func (cmd *Command) Decode(in []byte) error {
	cmd.Timestamp.Time = time.Now()

	for _, tok := range strings.Split(string(in), "|") {
		tok := strings.SplitN(tok, ":", 2)

		if len(tok) != 2 {
			continue
		}

		switch tok[0] {
		case "n", "name":
			cmd.Topic = tok[1]
		case "v", "val", "value":
			cmd.Value = tok[1]
		case "m", "meth", "method":
			switch tok[1] {
			case "sb", "subscr", "subscribe":
				cmd.Method = SUB
			case "s", "set":
				cmd.Method = PUB
			case "rel", "release":
				cmd.Method = USB
			case "g", "gf", "get", "getfull":
				cmd.Method = GET
			default:
				return fmt.Errorf("unknown method")
			}
		case "t", "time":
			if err := cmd.Timestamp.Parse(tok[1]); err != nil {
				return fmt.Errorf("handle time token: %v", err)
			}
		}
	}

	if cmd.Method == 0 {
		return fmt.Errorf("method not found")
	}

	if cmd.Topic == "" {
		return fmt.Errorf("topic not found.")
	}

	return nil
}

func (cmd *Command) Encode() []byte {
	value := "none"

	if cmd.Value != nil {
		value = fmt.Sprint(cmd.Value)
	}

	msg := fmt.Sprintf("time:%v|name:%v|descr:-|units:-|type:rw|val:%v\n",
		cmd.Timestamp.Format(),
		cmd.Topic,
		value,
	)

	return []byte(msg)
}
