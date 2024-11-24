package proc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeParse(t *testing.T) {
	var tm Time

	err := tm.Parse("11.06.2005 23_59_59.999")

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, int64(1118509199999), tm.UnixMilli())
}

func TestTimeParseError(t *testing.T) {
	var tm Time

	assert.NotNil(t, tm.Parse("11:06:2005 23_59_59.999"))
}

func TestTimeFormat(t *testing.T) {
	var tm Time

	tm.Set(1118509199999)

	assert.Equal(t, "11.06.2005 23_59_59.999", tm.Format())
}

func TestRecordMarshal(t *testing.T) {
	rec := Record{
		Value: "0x1118509199999",
	}

	rec.Timestamp.Set(1118509199999)
	res, err := json.Marshal(rec)

	assert.Nil(t, err)
	assert.Equal(t, `{"value":"0x1118509199999","timestamp":1118509199999}`, string(res))
}

func TestRecordMarshalEmptyValue(t *testing.T) {
	var rec Record

	rec.Timestamp.Set(1118509199999)
	res, err := json.Marshal(rec)

	assert.Nil(t, err)
	assert.Equal(t, `{"timestamp":1118509199999}`, string(res))
}

func TestRecordUnmarshal(t *testing.T) {
	var rec Record

	err := json.Unmarshal([]byte(`
	{
		"extra": 1118509199999,
		"value": "0x1118509199999",
		"timestamp": 1118509199999
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, "0x1118509199999", rec.Value)
	assert.Equal(t, int64(1118509199999), rec.Timestamp.UnixMilli())
}

func TestRecordUnmarshalEmptyValue(t *testing.T) {
	var rec Record

	rec.Value = "none"

	err := json.Unmarshal([]byte(`
	{
		"extra": 1118509199999,
		"timestamp": 1118509199999
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, "none", rec.Value)
	assert.Equal(t, int64(1118509199999), rec.Timestamp.UnixMilli())
}

func TestRecordUnmarshalEmptyTimestamp(t *testing.T) {
	var rec Record

	rec.Timestamp.Set(1118509199000)

	err := json.Unmarshal([]byte(`
	{
		"extra": 1118509199999,
		"value": "0x1118509199999"
	}
	`), &rec)

	assert.Nil(t, err)
	assert.Equal(t, "0x1118509199999", rec.Value)
	assert.Equal(t, int64(1118509199000), rec.Timestamp.UnixMilli())
}

func TestCommandDecodeShortPub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:s|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeShortSub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:sb|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, SUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeShortUsb(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:f|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, USB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

}

func TestCommandDecodeShortGet(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:g|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, GET, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

	assert.Nil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:gf|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, GET, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeMidPub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:pub|val:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeMidSub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:sub|val:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, SUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

	assert.Nil(t, cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:subscr|val:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, SUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeMidUsb(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("tm:11.06.2005 23_59_59.999|nm:test|meth:rel|val:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, USB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeLongPub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:set|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:publish|v:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, PUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeLongSub(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:subscribe|value:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, SUB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeLongUsb(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:free|value:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, USB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:release|value:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, USB, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func TestCommandDecodeLongGet(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:getfull|value:11.06")))
	assert.Equal(t, "test", cmd.Topic)
	assert.Equal(t, GET, cmd.Method)
	assert.Equal(t, "11.06", cmd.Value)
	assert.Equal(t, int64(1118509199999), cmd.Timestamp.UnixMilli())
}

func BenchmarkCommandDecode(b *testing.B) {
	var cmd Command

	for range b.N {
		cmd.Decode([]byte("time:11.06.2005 23_59_59.999|name:test|method:getfull|value:11.06"))
	}
}

func TestCommandDecodeUnknownMethodError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|m:unknown|v:11.06")))
}

func TestCommandDecodeMethodMissedError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|n:test|v:11.06")))
}

func TestCommandDecodeTopicMissedError(t *testing.T) {
	var cmd Command

	assert.NotNil(t, cmd.Decode([]byte("t:11.06.2005 23_59_59.999|m:set|v:11.06")))
}

func TestCommandDecodeTimeMissed(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("n:test|m:set|v:11.06")))
	assert.Less(t, time.Since(cmd.Timestamp.Time), 3*time.Second)
}

func TestCommandDecodeExtraToken(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("n:test|extra:token|m:set|v:11.06")))
}

func TestCommandDecodeMalformedToken(t *testing.T) {
	var cmd Command

	assert.Nil(t, cmd.Decode([]byte("n:test|malformed|m:set|v:11.06")))
}

func TestCommandEncode(t *testing.T) {
	cmd := Command{
		Topic: "test",
		Record: Record{
			Value:     "11.06",
			Timestamp: UnixMilli(1118509199999),
		},
	}

	enc, err := cmd.Encode()

	assert.Nil(t, err)
	assert.Equal(t, "time:11.06.2005 23_59_59.999|name:test|descr:-|units:-|type:rw|val:11.06\n", string(enc))
}

func BenchmarkCommandEncode(b *testing.B) {
	cmd := Command{
		Topic: "test",
		Record: Record{
			Value:     "11.06",
			Timestamp: UnixMilli(1118509199999),
		},
	}

	for range b.N {
		cmd.Encode()
	}
}

func TestCommandEncodeValueEmpty(t *testing.T) {
	cmd := Command{
		Topic: "test",
		Record: Record{
			Timestamp: UnixMilli(1118509199999),
		},
	}

	enc, err := cmd.Encode()

	assert.Nil(t, err)
	assert.Equal(t, "time:11.06.2005 23_59_59.999|name:test|descr:-|units:-|type:rw|val:none\n", string(enc))
}
