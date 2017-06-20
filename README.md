logrus_kinesis
====

[![Build Status](https://travis-ci.org/evalphobia/logrus_kinesis.svg?branch=master)](https://travis-ci.org/evalphobia/logrus_kinesis) [![Coverage Status](https://coveralls.io/repos/evalphobia/logrus_kinesis/badge.svg?branch=master&service=github)](https://coveralls.io/github/evalphobia/logrus_kinesis?branch=master) [![codecov](https://codecov.io/gh/evalphobia/logrus_kinesis/branch/master/graph/badge.svg)](https://codecov.io/gh/evalphobia/logrus_kinesis)
 [![GoDoc](https://godoc.org/github.com/evalphobia/logrus_kinesis?status.svg)](https://godoc.org/github.com/evalphobia/logrus_kinesis)


# AWS Kinesis Hook for Logrus <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:"/>

## Usage

```go
import (
    "github.com/evalphobia/logrus_kinesis"
    "github.com/sirupsen/logrus"
)

func main() {
    hook, err := logrus_kinesis.New("my_stream", Config{
        AccessKey: "ABC", // AWS accessKeyId
        SecretKey: "XYZ", // AWS secretAccessKey
        Region:    "ap-northeast-1",
    })

    // set custom fire level
    hook.SetLevels([]logrus.Level{
        logrus.PanicLevel,
        logrus.ErrorLevel,
    })

    // ignore field
    hook.AddIgnore("context")

    // add custome filter
    hook.AddFilter("error", logrus_kinesis.FilterError)


    // send log with logrus
    logger := logrus.New()
    logger.Hooks.Add(hook)
    logger.WithFields(f).Error("my_message") // send log data to kinesis as JSON
}
```


## Special fields

Some logrus fields have a special meaning in this hook.

|||
|:--|:--|
|`message`|if `message` is not set, entry.Message is added to log data in "message" field. |
|`stream_name`|`stream_name` is a custom stream name for Kinesis. If not set, `defaultStreamName` is used as stream name.|
|`partition_key`|`partition_key` is a custom partition key for Kinesis. If not set, `defaultStreamName` or entry.Message is used as stream name.|
