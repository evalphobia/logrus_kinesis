package logrus_kinesis

import (
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)
	return

	hook, err := New("test_stream", Config{})
	assert.Error(err)
	assert.Nil(hook)
}

func TestNewWithAWSConfig(t *testing.T) {
	assert := assert.New(t)
	return

	hook, err := NewWithAWSConfig("test_stream", nil)
	assert.Error(err)
	assert.Nil(hook)
}

func TestLevels(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		levels []logrus.Level
	}{
		{nil},
		{[]logrus.Level{logrus.WarnLevel}},
		{[]logrus.Level{logrus.ErrorLevel}},
		{[]logrus.Level{logrus.WarnLevel, logrus.DebugLevel}},
		{[]logrus.Level{logrus.WarnLevel, logrus.DebugLevel, logrus.ErrorLevel}},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		hook := KinesisHook{}
		levels := hook.Levels()
		assert.Nil(levels, target)

		hook.levels = tt.levels
		levels = hook.Levels()
		assert.Equal(tt.levels, levels, target)
	}
}

func TestSetLevels(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		levels []logrus.Level
	}{
		{nil},
		{[]logrus.Level{logrus.WarnLevel}},
		{[]logrus.Level{logrus.ErrorLevel}},
		{[]logrus.Level{logrus.WarnLevel, logrus.DebugLevel}},
		{[]logrus.Level{logrus.WarnLevel, logrus.DebugLevel, logrus.ErrorLevel}},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		hook := KinesisHook{}
		assert.Nil(hook.levels, target)

		hook.SetLevels(tt.levels)
		assert.Equal(tt.levels, hook.levels, target)

		hook.SetLevels(nil)
		assert.Nil(hook.levels, target)
	}
}

func TestAddIgnore(t *testing.T) {
	assert := assert.New(t)

	hook := KinesisHook{
		ignoreFields: make(map[string]struct{}),
	}

	list := []string{"foo", "bar", "baz"}
	for i, key := range list {
		assert.Len(hook.ignoreFields, i)

		hook.AddIgnore(key)
		assert.Len(hook.ignoreFields, i+1)

		for j := 0; j <= i; j++ {
			assert.Contains(hook.ignoreFields, list[j])
		}
	}
}

func TestAddFilter(t *testing.T) {
	assert := assert.New(t)

	hook := KinesisHook{
		filters: make(map[string]func(interface{}) interface{}),
	}

	list := []string{"foo", "bar", "baz"}
	for i, key := range list {
		assert.Len(hook.filters, i)

		hook.AddFilter(key, nil)
		assert.Len(hook.filters, i+1)

		for j := 0; j <= i; j++ {
			assert.Contains(hook.filters, list[j])
		}
	}
}

func TestGetStreamName(t *testing.T) {
	assert := assert.New(t)

	emptyEntry := ""
	tests := []struct {
		hasEntryName bool
		entryName    interface{}
		defautName   string
		expectedName string
	}{
		{true, "entry_stream", "default_stream", "entry_stream"},
		{true, "entry_stream", "", "entry_stream"},
		{true, "", "default_stream", ""},
		{true, "", "", ""},
		{true, 99999, "default_stream", "default_stream"},
		{true, nil, "default_stream", "default_stream"},
		{false, emptyEntry, "default_stream", "default_stream"},
		{false, emptyEntry, "", ""},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		hook := KinesisHook{
			defaultStreamName: tt.defautName,
		}
		entry := &logrus.Entry{
			Data: make(map[string]interface{}),
		}
		if tt.hasEntryName {
			entry.Data["stream_name"] = tt.entryName
		}

		assert.Equal(tt.expectedName, hook.getStreamName(entry), target)
	}
}

func TestPartitionKey(t *testing.T) {
	assert := assert.New(t)

	emptyEntry := ""
	tests := []struct {
		hasEntryKey bool
		entryKey    interface{}
		defautKey   string
		message     string
		expectedKey string
	}{
		{true, "entry_key", "default_key", "message_key", "entry_key"},
		{true, "entry_key", "default_key", "", "entry_key"},
		{true, "entry_key", "", "message_key", "entry_key"},
		{true, "entry_key", "", "", "entry_key"},
		{true, "", "default_key", "message_key", ""},
		{true, "", "default_key", "", ""},
		{true, "", "", "message_key", ""},
		{true, "", "", "", ""},
		{true, 99999, "default_key", "message_key", "default_key"},
		{true, nil, "default_key", "message_key", "default_key"},
		{true, 99999, "default_key", "", "default_key"},
		{true, nil, "default_key", "", "default_key"},
		{true, 99999, "", "message_key", "message_key"},
		{true, nil, "", "message_key", "message_key"},
		{false, emptyEntry, "default_key", "message_key", "default_key"},
		{false, emptyEntry, "default_key", "", "default_key"},
		{false, emptyEntry, "", "message_key", "message_key"},
		{false, emptyEntry, "", "", ""},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		hook := KinesisHook{
			defaultPartitionKey: tt.defautKey,
		}
		entry := &logrus.Entry{
			Message: tt.message,
			Data:    make(map[string]interface{}),
		}
		if tt.hasEntryKey {
			entry.Data["partition_key"] = tt.entryKey
		}

		assert.Equal(tt.expectedKey, hook.getPartitionKey(entry), target)
	}
}

func TestGetData(t *testing.T) {
	assert := assert.New(t)

	const defaultMessage = "entry_message"

	tests := []struct {
		data     map[string]interface{}
		expected string
	}{
		{
			map[string]interface{}{},
			`{"message":"entry_message"}`,
		},
		{
			map[string]interface{}{"message": "field_message"},
			`{"message":"field_message"}`,
		},
		{
			map[string]interface{}{
				"name":  "apple",
				"price": 105,
				"color": "red",
			},
			`{"color":"red","message":"entry_message","name":"apple","price":105}`,
		},
		{
			map[string]interface{}{
				"name":    "apple",
				"price":   105,
				"color":   "red",
				"message": "field_message",
			},
			`{"color":"red","message":"field_message","name":"apple","price":105}`,
		},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		hook := KinesisHook{}
		entry := &logrus.Entry{
			Message: defaultMessage,
			Data:    tt.data,
		}

		assert.Equal(tt.expected, string(hook.getData(entry)), target)
	}
}
func TestStringPtr(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		value string
	}{
		{"abc"},
		{""},
		{"991029102910291029478748"},
		{"skjdklsajdlewrjo4iuoivjcklxmc,.mklrjtlkrejijoijpoijvpodjfr"},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		p := stringPtr(tt.value)
		assert.Equal(tt.value, *p, target)
	}
}

// func TestFire(t *testing.T) {
// 	return
// 	hook, err := New("test_stream", Config{
// 		AccessKey: "",
// 		SecretKey: "",
// 		Region:    "ap-northeast-1",
// 	})
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	}
// 	logrus.AddHook(hook)

// 	logger := logrus.New()
// 	logger.Hooks.Add(hook)

// 	f := logrus.Fields{
// 		"message?": "fieldMessage",
// 		"tag":      "fieldTag",
// 		"value":    "fieldValue",
// 	}

// 	logger.WithFields(f).Error("my_message")

// 	return
// }
