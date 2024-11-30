package vcas

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	Stamp = "02.01.2006 15_04_05.000"

	PUB Method = iota + 1
	SUB
	USB
	GET
)

type Method int

func (m *Method) MarshalText() ([]byte, error) {
	var s string

	switch *m {
	case PUB:
		s = "set"
	case SUB:
		s = "subscribe"
	case USB:
		s = "release"
	case GET:
		s = "get"
	default:
		return nil, fmt.Errorf("unknown method id: %v", *m)
	}

	return []byte(s), nil
}

func (m *Method) UnmarshalText(b []byte) error {
	s := string(b)

	switch s {
	case "s", "set":
		*m = PUB
	case "sb", "subscr", "subscribe":
		*m = SUB
	case "rel", "release":
		*m = USB
	case "g", "gf", "get", "getfull":
		*m = GET
	default:
		return fmt.Errorf("unknown method: %v", s)
	}

	return nil
}

type Time struct {
	time.Time
}

func (t *Time) MarshalText() ([]byte, error) {
	return []byte(t.Format(Stamp)), nil
}

func (t *Time) UnmarshalText(b []byte) error {
	tm, err := time.ParseInLocation(Stamp, string(b), time.Local)

	if err == nil {
		t.Time = tm
	}

	return err
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.UnixMilli(), 10)), nil
}

func (t *Time) UnmarshalJSON(json []byte) error {
	milli, err := strconv.ParseInt(string(json), 10, 64)

	if err != nil {
		return fmt.Errorf("could not parse time: %v", err)
	}

	t.Time = time.UnixMilli(milli)

	return nil
}

func Unmarshal(b []byte, a any) error {
	t := reflect.TypeOf(a)

	if t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a struct pointer")
	}

	tok := make(map[string]string, 6)

	for _, t := range strings.Split(string(b), "|") {
		t := strings.SplitN(t, ":", 2)

		if len(t) != 2 {
			continue
		}

		tok[t[0]] = t[1]
	}

	v := reflect.ValueOf(a).Elem()
	t = t.Elem()

	for i := range t.NumField() {
		n, ok := t.Field(i).Tag.Lookup("vcas")

		if !ok {
			continue
		}

		f := v.Field(i)

		if !f.CanSet() {
			continue
		}

		for _, n := range strings.Split(n, ",") {
			v, ok := tok[n]

			if !ok {
				continue
			}

			switch f.Kind() {
			case reflect.Interface:
				f.Set(reflect.ValueOf(v))
			case reflect.String:
				f.SetString(v)
			case reflect.Float64:
				if p, err := strconv.ParseFloat(v, 64); err == nil {
					f.SetFloat(p)
				}
			case reflect.Int64:
				if p, err := strconv.ParseInt(v, 10, 64); err == nil {
					f.SetInt(p)
				}
			default:
				m, ok := f.Addr().Interface().(encoding.TextUnmarshaler)

				if !ok {
					return fmt.Errorf("unsupported type")
				}

				if err := m.UnmarshalText([]byte(v)); err != nil {
					return fmt.Errorf("could not parse field: %v", err)
				}
			}

			break
		}
	}

	return nil
}

func Marshal(a any) ([]byte, error) {
	var s strings.Builder

	s.Grow(76)
	s.WriteString("descr:-|type:rw|units:-")

	v := reflect.ValueOf(a).Elem()
	t := v.Type()

	for i := range t.NumField() {
		n, ok := t.Field(i).Tag.Lookup("vcas")

		if !ok {
			continue
		}

		f := v.Field(i)

		if !f.CanInterface() {
			continue
		}

		n = strings.Split(n, ",")[0]

		if m, ok := f.Addr().Interface().(encoding.TextMarshaler); ok {
			e, err := m.MarshalText()

			if err != nil {
				return nil, fmt.Errorf("could not encode value: %v", err)
			}

			s.WriteString(fmt.Sprintf("|%s:%v", n, string(e)))
		} else {
			s.WriteString(fmt.Sprintf("|%s:%v", n, f.Interface()))
		}
	}

	return []byte(s.String()), nil
}
