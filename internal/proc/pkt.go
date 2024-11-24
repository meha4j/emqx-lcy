package proc

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	Location *time.Location
)

const (
	Stamp = "02.01.2006 15_04_05.000"

	PUB = iota + 1
	SUB
	USB
	GET
)

func init() {
	loc, err := time.LoadLocation("Asia/Novosibirsk")

	if err != nil {
		panic(err)
	}

	Location = loc
}

type Time struct {
	time.Time
}

func Now() Time {
	return Time{
		time.Now().In(Location),
	}
}

func UnixMilli(milli int64) Time {
	return Time{time.UnixMilli(milli)}
}

func (t *Time) Set(milli int64) {
	t.Time = time.UnixMilli(milli).In(Location)
}

func (t *Time) Parse(value string) error {
	res, err := time.ParseInLocation(Stamp, value, Location)

	if err != nil {
		return err
	}

	t.Time = res

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
		return err
	}

	t.Set(milli)

	return nil
}

type Record struct {
	Value     string `json:"value,omitempty"`
	Timestamp Time   `json:"timestamp"`
}

type Command struct {
	Topic  string
	Method int

	Record
}

func (cmd *Command) Decode(in []byte) error {
	cmd.Timestamp = Now()
	tok := strings.Split(string(in), "|")

	for _, t := range tok {
		tok := strings.SplitN(t, ":", 2)

		if len(tok) != 2 {
			continue
		}

		switch tok[0] {
		case "n", "nm", "name":
			cmd.Topic = tok[1]
		case "v", "val", "value":
			cmd.Value = tok[1]
		case "t", "tm", "time":
			err := cmd.Timestamp.Parse(tok[1])

			if err != nil {
				return err
			}
		case "m", "meth", "method":
			switch tok[1] {
			case "sb", "sub", "subscr", "subscribe":
				cmd.Method = SUB
			case "s", "pub", "set", "publish":
				cmd.Method = PUB
			case "f", "rel", "free", "release":
				cmd.Method = USB
			case "g", "gf", "get", "getfull":
				cmd.Method = GET
			default:
				return fmt.Errorf("Malformed packet: method unknown: %v.", tok[1])
			}
		}
	}

	if cmd.Method == 0 {
		return fmt.Errorf("Malformed packet: method not found.")
	}

	if cmd.Topic == "" {
		return fmt.Errorf("Malformed packet: topic not found.")
	}

	return nil
}

func (cmd *Command) Encode() ([]byte, error) {
	value := "none"

	if cmd.Value != "" {
		value = cmd.Value
	}

	msg := fmt.Sprintf("time:%v|name:%v|descr:-|units:-|type:rw|val:%v\n",
		cmd.Timestamp.Format(),
		cmd.Topic,
		value,
	)

	return []byte(msg), nil
}
