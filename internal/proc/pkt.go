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
)

const (
	PUB = iota
	SUB = iota
	USB = iota
	GET = iota
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

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

func (t *Time) UnmarshalJSON(json []byte) error {
	milli, err := strconv.ParseInt(string(json), 10, 64)

	if err != nil {
		return err
	}

	t.Time = time.UnixMilli(milli).In(Location)

	return nil
}

type Record struct {
	Value string `json:"value,omitempty"`
	Stamp Time   `json:"timestamp"`
}

type Command struct {
	Topic  string
	Method int

	Record
}

func (cmd *Command) Decode(in []byte) error {
	cmd.Stamp.Time = time.Now().In(Location)
	cmd.Method = -1

	tok := strings.Split(string(in), "|")

	for _, t := range tok {
		tok := strings.Split(t, ":")

		if len(tok) != 2 {
			continue
		}

		switch tok[0] {
		case "n", "name":
			cmd.Topic = tok[1]
		case "v", "val", "value":
			cmd.Value = tok[1]
		case "t", "time":
			tm, err := time.ParseInLocation(Stamp, tok[1], Location)

			if err != nil {
				return err
			}

			cmd.Stamp.Time = tm
		case "m", "meth", "method":
			switch tok[1] {
			case "sb", "sub", "subscr", "subscribe":
				cmd.Method = SUB
			case "s", "set":
				cmd.Method = PUB
			case "f", "ref", "free", "release":
				cmd.Method = USB
			case "g", "gf", "get", "getfull":
				cmd.Method = GET
			default:
				return fmt.Errorf("Malformed packet: method unknown: %v.", tok[1])
			}
		}
	}

	if cmd.Method == -1 {
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
		cmd.Stamp.Format(Stamp),
		cmd.Topic,
		value,
	)

	return []byte(msg), nil
}
