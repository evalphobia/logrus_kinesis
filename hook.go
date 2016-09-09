package logrus_kinesis

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var defaultLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
	logrus.WarnLevel,
	logrus.InfoLevel,
}

// KinesisHook is logrus hook for AWS kinesis.
type KinesisHook struct {
	client *kinesis.Kinesis

	defaultStreamName   string
	defaultPartitionKey string
	async               bool
	levels              []logrus.Level
	ignoreFields        map[string]struct{}
	filters             map[string]func(interface{}) interface{}
}

// New returns initialized logrus hook for fluentd with persistent fluentd logger.
func New(name string, conf Config) (*KinesisHook, error) {
	sess, err := session.NewSession(conf.AWSConfig())
	if err != nil {
		return nil, err
	}

	svc := kinesis.New(sess)
	return &KinesisHook{
		client:            svc,
		defaultStreamName: name,
		levels:            defaultLevels,
		ignoreFields:      make(map[string]struct{}),
		filters:           make(map[string]func(interface{}) interface{}),
	}, nil
}

// NewWithConfig returns initialized logrus hook for fluentd with persistent fluentd logger.
func NewWithAWSConfig(name string, conf *aws.Config) (*KinesisHook, error) {
	sess, err := session.NewSession(conf)
	if err != nil {
		return nil, err
	}

	svc := kinesis.New(sess)
	return &KinesisHook{
		client:       svc,
		levels:       defaultLevels,
		ignoreFields: make(map[string]struct{}),
		filters:      make(map[string]func(interface{}) interface{}),
	}, nil
}

// Levels returns logging level to fire this hook.
func (h *KinesisHook) Levels() []logrus.Level {
	return h.levels
}

// SetLevels sets logging level to fire this hook.
func (h *KinesisHook) SetLevels(levels []logrus.Level) {
	h.levels = levels
}

// SetPartitionKey sets default partitionKey for kinesis.
func (h *KinesisHook) SetPartitionKey(key string) {
	h.defaultPartitionKey = key
}

// Async sets async flag and send log asynchroniously.
// If use this option, Fire() does not return error.
func (h *KinesisHook) Async() {
	h.async = true
}

// AddIgnore adds field name to ignore.
func (h *KinesisHook) AddIgnore(name string) {
	h.ignoreFields[name] = struct{}{}
}

// AddFilter adds a custom filter function.
func (h *KinesisHook) AddFilter(name string, fn func(interface{}) interface{}) {
	h.filters[name] = fn
}

// Fire is invoked by logrus and sends log to kinesis.
func (h *KinesisHook) Fire(entry *logrus.Entry) error {
	if !h.async {
		return h.fire(entry)
	}

	// send log asynchroniously and return no error.
	go h.fire(entry)
	return nil
}

// Fire is invoked by logrus and sends log to kinesis.
func (h *KinesisHook) fire(entry *logrus.Entry) error {
	in := &kinesis.PutRecordInput{
		StreamName:   stringPtr(h.getStreamName(entry)),
		PartitionKey: stringPtr(h.getPartitionKey(entry)),
		Data:         h.getData(entry),
	}
	_, err := h.client.PutRecord(in)
	return err
}

func (h *KinesisHook) getStreamName(entry *logrus.Entry) string {
	if name, ok := entry.Data["stream_name"].(string); ok {
		return name
	}
	return h.defaultStreamName
}

func (h *KinesisHook) getPartitionKey(entry *logrus.Entry) string {
	if key, ok := entry.Data["partition_key"].(string); ok {
		return key
	}
	if h.defaultPartitionKey != "" {
		return h.defaultPartitionKey
	}
	return entry.Message
}

func (h *KinesisHook) getData(entry *logrus.Entry) []byte {
	if _, ok := entry.Data["message"]; !ok {
		entry.Data["message"] = entry.Message
	}

	data := make(logrus.Fields)
	for k, v := range entry.Data {
		if _, ok := h.ignoreFields[k]; ok {
			continue
		}
		if fn, ok := h.filters[k]; ok {
			v = fn(v) // apply custom filter
		} else {
			v = formatData(v) // use default formatter
		}
		data[k] = v
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return bytes
}

// formatData returns value as a suitable format.
func formatData(value interface{}) (formatted interface{}) {
	switch value := value.(type) {
	case json.Marshaler:
		return value
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	default:
		return value
	}
}

func stringPtr(str string) *string {
	return &str
}
