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

func (m Method) MarshalText() ([]byte, error) {
	var s string

	switch m {
	case PUB:
		s = "set"
	case SUB:
		s = "subscribe"
	case USB:
		s = "release"
	case GET:
		s = "get"
	default:
		return nil, fmt.Errorf("unknown method id: %v", m)
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

func (t Time) MarshalText() ([]byte, error) {
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
		return fmt.Errorf("parse millis: %v", err)
	}

	t.Time = time.UnixMilli(milli)

	return nil
}

func Unmarshal(b []byte, a any) error {
	v := reflect.ValueOf(a)

	return unmarshal(b, &v)
}

func unmarshal(b []byte, v *reflect.Value) error {
	if v.Kind() == reflect.Pointer {
		e := v.Elem()

		return unmarshal(b, &e)
	}

	if v.CanInterface() {
		if t, ok := v.Interface().(encoding.TextUnmarshaler); ok {
			return t.UnmarshalText(b)
		}
	}

	if v.CanAddr() && v.Addr().CanInterface() {
		if t, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return t.UnmarshalText(b)
		}
	}

	if !v.CanSet() {
		return nil
	}

	m := string(b)

	switch v.Type().Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(m))
	case reflect.String:
		v.SetString(m)
	case reflect.Int:
		if p, err := strconv.ParseInt(m, 10, 0); err == nil {
			v.SetInt(p)
		}
	case reflect.Int8:
		if p, err := strconv.ParseInt(m, 10, 8); err == nil {
			v.SetInt(p)
		}
	case reflect.Int16:
		if p, err := strconv.ParseInt(m, 10, 16); err == nil {
			v.SetInt(p)
		}
	case reflect.Int32:
		if p, err := strconv.ParseInt(m, 10, 32); err == nil {
			v.SetInt(p)
		}
	case reflect.Int64:
		if p, err := strconv.ParseInt(m, 10, 64); err == nil {
			v.SetInt(p)
		}
	case reflect.Uint:
		if p, err := strconv.ParseUint(m, 10, 0); err == nil {
			v.SetUint(p)
		}
	case reflect.Uint8:
		if p, err := strconv.ParseUint(m, 10, 8); err == nil {
			v.SetUint(p)
		}
	case reflect.Uint16:
		if p, err := strconv.ParseUint(m, 10, 16); err == nil {
			v.SetUint(p)
		}
	case reflect.Uint32:
		if p, err := strconv.ParseUint(m, 10, 32); err == nil {
			v.SetUint(p)
		}
	case reflect.Uint64:
		if p, err := strconv.ParseUint(m, 10, 64); err == nil {
			v.SetUint(p)
		}
	case reflect.Float32:
		if p, err := strconv.ParseFloat(m, 32); err == nil {
			v.SetFloat(p)
		}
	case reflect.Float64:
		if p, err := strconv.ParseFloat(m, 64); err == nil {
			v.SetFloat(p)
		}
	case reflect.Complex64:
		if p, err := strconv.ParseComplex(m, 64); err == nil {
			v.SetComplex(p)
		}
	case reflect.Complex128:
		if p, err := strconv.ParseComplex(m, 128); err == nil {
			v.SetComplex(p)
		}
	case reflect.Bool:
		if p, err := strconv.ParseBool(m); err == nil {
			v.SetBool(p)
		}
	case reflect.Map:
		if err := unmarshalMap(parseMap(m), v); err != nil {
			return fmt.Errorf("map unmarshal: %v", err)
		}
	case reflect.Struct:
		if err := unmarshalStruct(parseMap(m), v); err != nil {
			return fmt.Errorf("struct unmarshal: %v", err)
		}
	default:
		return fmt.Errorf("unsupported type")
	}

	return nil
}

func unmarshalMap(tok map[string]string, v *reflect.Value) error {
	if v.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("unsupported key type")
	}

	for mk, mv := range tok {
		rv := reflect.New(v.Type().Elem())

		if err := unmarshal([]byte(mv), &rv); err != nil {
			return fmt.Errorf("value unmarshal: %v", err)
		}

		v.SetMapIndex(reflect.ValueOf(mk), rv.Elem())
	}

	return nil
}

func unmarshalStruct(tok map[string]string, v *reflect.Value) error {
	t := v.Type()

	for i := range t.NumField() {
		f := v.Field(i)
		n, ok := t.Field(i).Tag.Lookup("vcas")

		if !ok {
			switch f.Type().Kind() {
			case reflect.Struct:
				if err := unmarshalStruct(tok, &f); err != nil {
					return fmt.Errorf("embedded struct unmarshal: %v", err)
				}
			case reflect.Map:
				if err := unmarshalMap(tok, &f); err != nil {
					return fmt.Errorf("inner map unmarshal: %v", err)
				}
			}

			continue
		}

		for _, n := range strings.Split(n, ",") {
			if v, ok := tok[n]; ok {
				if err := unmarshal([]byte(v), &f); err != nil {
					return fmt.Errorf("field unmarshal: %v", err)
				}

				break
			}
		}
	}

	return nil
}

func parseMap(m string) map[string]string {
	tok := make(map[string]string, 6)

	for _, t := range strings.Split(m, "|") {
		t := strings.SplitN(t, ":", 2)

		if len(t) != 2 {
			continue
		}

		tok[t[0]] = t[1]
	}

	return tok
}

func Marshal(a any) ([]byte, error) {
	var b strings.Builder

	v := reflect.ValueOf(a)

	if err := marshal(&v, &b); err != nil {
		return nil, fmt.Errorf("object marshal: %v", err)
	}

	return []byte(b.String()), nil
}

func marshal(v *reflect.Value, b *strings.Builder) error {
	if v.Kind() == reflect.Pointer {
		e := v.Elem()

		return marshal(&e, b)
	}

	if v.CanInterface() {
		if t, ok := v.Interface().(encoding.TextMarshaler); ok {
			txt, err := t.MarshalText()

			if err != nil {
				return fmt.Errorf("custom marshal: %v", err)
			}

			b.Write(txt)

			return nil
		}
	}

	if v.CanAddr() && v.Addr().CanInterface() {
		if t, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
			txt, err := t.MarshalText()

			if err != nil {
				return fmt.Errorf("custom marshal: %v", err)
			}

			b.Write(txt)

			return nil
		}
	}

	switch v.Kind() {
	case reflect.Interface:
		b.Write([]byte(fmt.Sprint(v.Interface())))
	case reflect.String:
		b.WriteString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(fmt.Sprint(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(fmt.Sprint(v.Uint()))
	case reflect.Float32, reflect.Float64:
		b.WriteString(fmt.Sprint(v.Float()))
	case reflect.Complex64, reflect.Complex128:
		b.WriteString(fmt.Sprint(v.Complex()))
	case reflect.Bool:
		b.WriteString(fmt.Sprint(v.Bool()))
	case reflect.Map:
		if err := marshalMap(v, b); err != nil {
			return fmt.Errorf("map marshal: %v", err)
		}
	case reflect.Struct:
		if err := marshalStruct(v, b); err != nil {
			return fmt.Errorf("struct marshal: %v", err)
		}
	default:
		return fmt.Errorf("unsupported type")
	}

	return nil
}

func marshalMap(v *reflect.Value, b *strings.Builder) error {
	iter := v.MapRange()

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		if b.Len() != 0 {
			b.WriteRune('|')
		}

		if err := marshal(&key, b); err != nil {
			return fmt.Errorf("key marshal: %v", err)
		}

		b.WriteRune(':')

		if err := marshal(&val, b); err != nil {
			return fmt.Errorf("value marshal: %v", err)
		}
	}

	return nil
}

func marshalStruct(v *reflect.Value, b *strings.Builder) error {
	t := v.Type()

	for i := range t.NumField() {
		n, ok := t.Field(i).Tag.Lookup("vcas")
		f := v.Field(i)

		if !ok {
			switch f.Type().Kind() {
			case reflect.Struct:
				if err := marshalStruct(&f, b); err != nil {
					return fmt.Errorf("embedded struct marshal: %v", err)
				}
			case reflect.Map:
				if err := marshalMap(&f, b); err != nil {
					return fmt.Errorf("inner map marshal: %v", err)
				}
			}

			continue
		}

		n = strings.Split(n, ",")[0]

		if b.Len() != 0 {
			b.WriteRune('|')
		}

		b.WriteString(n)
		b.WriteRune(':')

		if err := marshal(&f, b); err != nil {
			return fmt.Errorf("field marshal: %v", err)
		}
	}

	return nil
}
