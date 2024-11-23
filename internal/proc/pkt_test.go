package proc

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func check(t *testing.T, got any, want any) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`got %v, want %v`, got, want)
	}
}

var tm Time

func TestTimeParse(t *testing.T) {
	err := tm.Parse("11.06.2005 23_59_59.999")

	if err != nil {
		t.Error(err)
	}

	check(t, tm.UnixMilli(), int64(1118509199999))
}

func TestTimeParseError(t *testing.T) {
	err := tm.Parse("11:06:2005 23_59_59.999")

	if err == nil {
		t.Fail()
	}
}

func TestTimeFormat(t *testing.T) {
	tm.Time = time.UnixMilli(1118509199999)

	if tm.Format() != "11.06.2005 23_59_59.999" {
		t.Fail()
	}
}

var rec Record

func TestRecordMarshal(t *testing.T) {
	rec.Timestamp.Set(1118509199999)
	rec.Value = "0x1118509199999"

	res, err := json.Marshal(rec)

	if err != nil {
		t.Error(err)
	}

	check(t, string(res), `{"value":"0x1118509199999","timestamp":1118509199999}`)
}

func TestRecordMarshalEmptyValue(t *testing.T) {
	rec.Value = ""
	rec.Timestamp.Set(1118509199999)

	res, err := json.Marshal(rec)

	if err != nil {
		t.Error(err)
	}

	check(t, string(res), `{"timestamp":1118509199999}`)
}

func TestRecordUnmarshal(t *testing.T) {
	err := json.Unmarshal([]byte(`
	{
		"extra": 1118509199999,
		"value": "0x1118509199999",
		"timestamp": 1118509199999
	}
	`), &rec)

	if err != nil {
		t.Error(err)
	}

	check(t, rec.Value, "0x1118509199999")
	check(t, rec.Timestamp.UnixMilli(), int64(1118509199999))
}

func TestRecordUnmarshalEmptyValue(t *testing.T) {
	rec.Value = "none"

	err := json.Unmarshal([]byte(`
	{
		"extra": 1118509199999,
		"timestamp": 1118509199999
	}
	`), &rec)

	if err != nil {
		t.Error(err)
	}

	check(t, rec.Value, "none")
	check(t, rec.Timestamp.UnixMilli(), int64(1118509199999))
}

func TestRecordUnmarshalEmptyTimestamp(t *testing.T) {
	rec.Timestamp.Set(1118509199000)

	err := json.Unmarshal([]byte(`
	{
		"extra": 1118509199999,
		"value": "0x1118509199999"
	}
	`), &rec)

	if err != nil {
		t.Error(err)
	}

	check(t, rec.Value, "0x1118509199999")
	check(t, rec.Timestamp.UnixMilli(), int64(1118509199000))
}

var cmd Command

func TestCommandDecodeShortPub(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:s|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Topic, "test")
	check(t, cmd.Method, PUB)
	check(t, cmd.Value, "11.06")
	check(t, cmd.Timestamp.UnixMilli(), int64(1118509199999))
}

func TestCommandDecodeShortSub(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:sb|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, SUB)
}

func TestCommandDecodeShortUsb(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:f|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, USB)

}

func TestCommandDecodeShortGet(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:g|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	if cmd.Method != GET {
		t.Fail()
	}

	err = cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:gf|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	if cmd.Method != GET {
		t.Fail()
	}
}

func TestCommandDecodeMidPub(t *testing.T) {
	err := cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:pub|val:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Topic, "test")
	check(t, cmd.Method, PUB)
	check(t, cmd.Value, "11.06")
	check(t, cmd.Timestamp.UnixMilli(), int64(1118509199999))
}

func TestCommandDecodeMidSub(t *testing.T) {
	err := cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:sub|val:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, SUB)

	err = cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:subscr|val:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, SUB)
}

func TestCommandDecodeMidUsb(t *testing.T) {
	err := cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:rel|val:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, USB)
}

func TestCommandDecodeLongPub(t *testing.T) {
	err := cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:set|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Topic, "test")
	check(t, cmd.Method, PUB)
	check(t, cmd.Value, "11.06")
	check(t, cmd.Timestamp.UnixMilli(), int64(1118509199999))

	err = cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:publish|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, PUB)
}

func TestCommandDecodeLongSub(t *testing.T) {
	err := cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:subscribe|value:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, SUB)
}

func TestCommandDecodeLongUsb(t *testing.T) {
	err := cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:free|value:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, USB)

	err = cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:release|value:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, USB)
}

func TestCommandDecodeLongGet(t *testing.T) {
	err := cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:getfull|value:11.06"))

	if err != nil {
		t.Error(err)
	}

	check(t, cmd.Method, GET)
}

func TestCommandDecodeUnknownMethodError(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:unknown|v:11.06"))

	if err == nil {
		t.Fail()
	}
}

func TestCommandDecodeMethodMissedError(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|v:11.06"))

	if err == nil {
		t.Fail()
	}
}

func TestCommandDecodeTopicMissedError(t *testing.T) {
	err := cmd.Decode([]byte("t:11.06.2005 23_59_59.999|m:set|v:11.06"))

	if err == nil {
		t.Fail()
	}
}

func TestCommandDecodeTimeMissed(t *testing.T) {
	err := cmd.Decode([]byte("n:test|m:set|v:11.06"))

	if err != nil {
		t.Error(err)
	}

	if time.Since(cmd.Timestamp.Time) > 3*time.Second {
		t.Fail()
	}
}

func TestCommandDecodeExtraToken(t *testing.T) {
	err := cmd.Decode([]byte("n:test|extra:token|m:set|v:11.06"))

	if err != nil {
		t.Error(err)
	}
}

func TestCommandDecodeMalformedToken(t *testing.T) {
	err := cmd.Decode([]byte("n:test|malformed|m:set|v:11.06"))

	if err != nil {
		t.Error(err)
	}
}

func TestCommandEncode(t *testing.T) {
	cmd.Topic = "test"
	cmd.Value = "11.06"
	cmd.Timestamp.Set(1118509199999)

	enc, err := cmd.Encode()

	if err != nil {
		t.Error(err)
	}

	check(t, string(enc), "time:11.06.2005 23_59_59.999|name:test|descr:-|units:-|type:rw|val:11.06\n")
}

func TestCommandEncodeValueEmpty(t *testing.T) {
	cmd.Topic = "test"
	cmd.Value = ""
	cmd.Timestamp.Set(1118509199999)

	enc, err := cmd.Encode()

	if err != nil {
		t.Error(err)
	}

	check(t, string(enc), "time:11.06.2005 23_59_59.999|name:test|descr:-|units:-|type:rw|val:none\n")
}
