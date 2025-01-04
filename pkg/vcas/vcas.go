// Package vcas implements encoding and decoding of VCAS.
// The mapping between VCAS and Go values is described
// in the documentation for the Marshal and Unmarshal functions.
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
	Name    = "VCAS"
	Version = "2.0.0-SNAPSHOT"

	Stamp = "02.01.2006 15_04_05.000"
	None  = "none"

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
	tm, err := time.Parse(Stamp, string(b))

	if err != nil {
		return fmt.Errorf("parse stamp: %v", err)
	}

	t.Time = tm

	return nil
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

// Unmarshal parses the VCAS-encoded data and stores the result
// in the value pointed by a.
//
// Unmarshal traverses the value a recursively.
//
// If an encountered value implements [TextUmarshaler] and is
// not a nil pointer, Unmarshal calls [TextUnmarshaler.UnmarshalText].
//
// If no [TextUnmarshaler.UnmarshalText] method is present and the
// value is one of the primitive types, Unmarshal calls corresponding
// method from [strconv] package.
//
// Otherwise, Unmarshal uses the following type-dependent encodings:
//
// To unmarshal VCAS into a struct (map), Unmarshal matches incoming object
// tags (keys) to the keys from parsed bytes. If there are several tags provided,
// the first matching one is selected.
//
// To unmarshal VCAS into a slice, Unmarshal traverses tokens recursively.
func Unmarshal(b []byte, a any) error {
	v := reflect.ValueOf(a)

	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("unsupported target")
	}

	return unmarshal(b, &v)
}

func unmarshal(b []byte, v *reflect.Value) error {
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && !v.IsNil() {
		e := v.Elem()

		return unmarshal(b, &e)
	}

	if v.CanInterface() {
		if t, ok := v.Interface().(encoding.TextUnmarshaler); ok {
			if err := t.UnmarshalText(b); err != nil {
				return fmt.Errorf("provided: %v", err)
			}

			return nil
		}
	}

	if v.CanAddr() && v.Addr().CanInterface() {
		if t, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := t.UnmarshalText(b); err != nil {
				return fmt.Errorf("provided: %v", err)
			}

			return nil
		}
	}

	if !v.CanSet() {
		return nil
	}

	msg := string(b)

	switch v.Type().Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(msg))
	case reflect.String:
		v.SetString(msg)
	case reflect.Int:
		if p, err := strconv.ParseInt(msg, 10, 0); err == nil {
			v.SetInt(p)
		} else {
			return err
		}
	case reflect.Int8:
		if p, err := strconv.ParseInt(msg, 10, 8); err == nil {
			v.SetInt(p)
		} else {
			return err
		}
	case reflect.Int16:
		if p, err := strconv.ParseInt(msg, 10, 16); err == nil {
			v.SetInt(p)
		} else {
			return err
		}
	case reflect.Int32:
		if p, err := strconv.ParseInt(msg, 10, 32); err == nil {
			v.SetInt(p)
		} else {
			return err
		}
	case reflect.Int64:
		if p, err := strconv.ParseInt(msg, 10, 64); err == nil {
			v.SetInt(p)
		} else {
			return err
		}
	case reflect.Uint:
		if p, err := strconv.ParseUint(msg, 10, 0); err == nil {
			v.SetUint(p)
		} else {
			return err
		}
	case reflect.Uint8:
		if p, err := strconv.ParseUint(msg, 10, 8); err == nil {
			v.SetUint(p)
		} else {
			return err
		}
	case reflect.Uint16:
		if p, err := strconv.ParseUint(msg, 10, 16); err == nil {
			v.SetUint(p)
		} else {
			return err
		}
	case reflect.Uint32:
		if p, err := strconv.ParseUint(msg, 10, 32); err == nil {
			v.SetUint(p)
		} else {
			return err
		}
	case reflect.Uint64:
		if p, err := strconv.ParseUint(msg, 10, 64); err == nil {
			v.SetUint(p)
		} else {
			return err
		}
	case reflect.Float32:
		if p, err := strconv.ParseFloat(msg, 32); err == nil {
			v.SetFloat(p)
		} else {
			return err
		}
	case reflect.Float64:
		if p, err := strconv.ParseFloat(msg, 64); err == nil {
			v.SetFloat(p)
		} else {
			return err
		}
	case reflect.Complex64:
		if p, err := strconv.ParseComplex(msg, 64); err == nil {
			v.SetComplex(p)
		} else {
			return err
		}
	case reflect.Complex128:
		if p, err := strconv.ParseComplex(msg, 128); err == nil {
			v.SetComplex(p)
		} else {
			return err
		}
	case reflect.Bool:
		if p, err := strconv.ParseBool(msg); err == nil {
			v.SetBool(p)
		} else {
			return err
		}
	case reflect.Map:
		if err := unmarshalMap(parseMap(msg), v); err != nil {
			return err
		}
	case reflect.Slice:
		if err := unmarshalSlice(parseSlice(msg), v); err != nil {
			return err
		}
	case reflect.Struct:
		if err := unmarshalStruct(parseMap(msg), v); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported type")
	}

	return nil
}

func unmarshalMap(tok map[string]string, val *reflect.Value) error {
	if val.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("unsupported key type")
	}

	for k, v := range tok {
		e := reflect.New(val.Type().Elem())

		if err := unmarshal([]byte(v), &e); err != nil {
			return fmt.Errorf("token (%s): %v", k, err)
		}

		val.SetMapIndex(reflect.ValueOf(k), e.Elem())
	}

	return nil
}

func unmarshalSlice(tok []string, val *reflect.Value) error {
	for i, v := range tok {
		e := reflect.New(val.Type().Elem())

		if err := unmarshal([]byte(v), &e); err != nil {
			return fmt.Errorf("token (%d): %v", i, err)
		}

		val.Set(reflect.Append(*val, e.Elem()))
	}

	return nil
}

func unmarshalStruct(tok map[string]string, val *reflect.Value) error {
	t := val.Type()

fs:
	for i := range t.NumField() {
		sf := val.Field(i)
		tf := t.Field(i)

		if ns, ok := tf.Tag.Lookup("vcas"); ok {
			for _, n := range strings.Split(ns, ",") {
				if v, ok := tok[n]; ok {
					if err := unmarshal([]byte(v), &sf); err != nil {
						return fmt.Errorf("token (%s): %v", tf.Name, err)
					}

					continue fs
				}
			}
		}

		if sf.Type().Kind() == reflect.Struct {
			if err := unmarshalStruct(tok, &sf); err != nil {
				return fmt.Errorf("token (%s): %v", tf.Name, err)
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

func parseSlice(m string) []string {
	return strings.Split(m, ",")
}

// Marshal returns the VCAS encoding of a.
//
// Unmarshal traverses the value a recursively.
//
// If an encountered value implements [TextMarshaler] and is
// not a nil pointer, Marshal calls [TextMarshaler.MarshalText].
//
// If no [TextMarshaler.MarshalText] method is present and the
// value is one of the primitive types, Marshal calls corresponding
// method from [fmt] package to encode value as string.
//
// Otherwise, to marshal struct, map or slice, Marshal traverses tokens recursively,
// selecting struct tags as keys. If there are multiple tags provided,
// the first one selected.
func Marshal(a any) ([]byte, error) {
	var b strings.Builder

	v := reflect.ValueOf(a)

	if err := marshal(&v, &b); err != nil {
		return nil, err
	}

	return []byte(b.String()), nil
}

func marshal(v *reflect.Value, b *strings.Builder) error {
	if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && !v.IsNil() {
		e := v.Elem()

		return marshal(&e, b)
	}

	if v.CanInterface() {
		if t, ok := v.Interface().(encoding.TextMarshaler); ok {
			txt, err := t.MarshalText()

			if err != nil {
				return fmt.Errorf("provided: %v", err)
			}

			b.Write(txt)

			return nil
		}
	}

	if v.CanAddr() && v.Addr().CanInterface() {
		if t, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
			txt, err := t.MarshalText()

			if err != nil {
				return fmt.Errorf("provided: %v", err)
			}

			b.Write(txt)

			return nil
		}
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		b.WriteString(None)
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
			return err
		}
	case reflect.Slice:
		if err := marshalSlice(v, b); err != nil {
			return err
		}
	case reflect.Struct:
		if err := marshalStruct(v, b); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported type")
	}

	return nil
}

func marshalMap(v *reflect.Value, b *strings.Builder) error {
	itr := v.MapRange()

	for itr.Next() {
		key := itr.Key()
		val := itr.Value()

		if b.Len() != 0 {
			b.WriteRune('|')
		}

		if err := marshal(&key, b); err != nil {
			return fmt.Errorf("token (%s): key: %v", key.String(), err)
		}

		b.WriteRune(':')

		if err := marshal(&val, b); err != nil {
			return fmt.Errorf("token (%s): value: %v", key.String(), err)
		}
	}

	return nil
}

func marshalSlice(v *reflect.Value, b *strings.Builder) error {
	for i := range v.Len() {
		if i != 0 {
			b.WriteRune(',')
		}

		e := v.Index(i)

		if err := marshal(&e, b); err != nil {
			return fmt.Errorf("token (%d): %v", i, err)
		}
	}

	return nil
}

func marshalStruct(v *reflect.Value, b *strings.Builder) error {
	t := v.Type()

	for i := range t.NumField() {
		sf := v.Field(i)
		tf := t.Field(i)

		if a, ok := tf.Tag.Lookup("vcas"); ok {
			a = strings.Split(a, ",")[0]

			if b.Len() != 0 {
				b.WriteRune('|')
			}

			b.WriteString(a)
			b.WriteRune(':')

			if err := marshal(&sf, b); err != nil {
				return fmt.Errorf("token (%s): %v", tf.Name, err)
			}

			continue
		}

		switch sf.Type().Kind() {
		case reflect.Map:
			if err := marshalMap(&sf, b); err != nil {
				return fmt.Errorf("token (%s): %v", tf.Name, err)
			}
		case reflect.Struct:
			if err := marshalStruct(&sf, b); err != nil {
				return fmt.Errorf("token (%s): %v", tf.Name, err)
			}
		}
	}

	return nil
}
