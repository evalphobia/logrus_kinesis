package logrus_kinesis

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSConfig(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		isSuccess bool
		provider  string
		accessKey string
		secretKey string
		region    string
	}{
		{true, "StaticProvider", "access_key", "secret_key", "region"},
		{true, "StaticProvider", "access_key", "secret_key", ""},
		// {true, "SharedCredentialsProvider", "", "", "region"},
	}

	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)

		conf := Config{
			AccessKey: tt.accessKey,
			SecretKey: tt.secretKey,
			Region:    tt.region,
		}
		awsConf := conf.AWSConfig()
		assert.NotNil(awsConf, target)
		val, err := awsConf.Credentials.Get()
		if !tt.isSuccess {
			assert.Error(err, target)
			return
		}

		assert.Equal(tt.provider, val.ProviderName, target)
		assert.Equal(tt.accessKey, val.AccessKeyID, target)
		assert.Equal(tt.secretKey, val.SecretAccessKey, target)

		if tt.region == "" {
			tt.region = defaultRegion
		}
		assert.Equal(tt.region, *awsConf.Region, target)
	}
}

func TestAWSCredentials(t *testing.T) {
	assert := assert.New(t)

	const useEnv = true
	const noEnv = false

	tests := []struct {
		isSuccess bool
		useEnv    bool
		provider  string
		accessKey string
		secretKey string
	}{
		{true, useEnv, "EnvProvider", "access_key", "secret_key"},
		{true, noEnv, "StaticProvider", "access_key", "secret_key"},
		{false, useEnv, "", "access_key", ""},
		{false, noEnv, "", "access_key", ""},
		// {true, noEnv, "SharedCredentialsProvider", "", ""},
	}

	defer os.Clearenv()
	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)
		os.Clearenv()

		conf := Config{}
		if tt.useEnv {
			os.Setenv("AWS_ACCESS_KEY_ID", tt.accessKey)
			os.Setenv("AWS_SECRET_ACCESS_KEY", tt.secretKey)
		} else {
			conf.AccessKey = tt.accessKey
			conf.SecretKey = tt.secretKey
		}

		cred := conf.awsCredentials()
		assert.NotNil(cred, target)
		val, err := cred.Get()
		if !tt.isSuccess {
			assert.Error(err, target)
			return
		}

		assert.Equal(tt.provider, val.ProviderName, target)
		assert.Equal(tt.accessKey, val.AccessKeyID, target)
		assert.Equal(tt.secretKey, val.SecretAccessKey, target)
	}
}

func TestGetRegion(t *testing.T) {
	assert := assert.New(t)

	os.Clearenv()
	defer os.Clearenv()

	conf := Config{}

	assert.Equal(defaultRegion, conf.getRegion(), "empty config, empty env")

	os.Setenv("AWS_REGION", "env_region")
	assert.Equal("env_region", conf.getRegion(), "empty config, set env")

	conf.Region = "conf_region"
	assert.Equal("conf_region", conf.getRegion(), "set config, set env")

	os.Clearenv()
	assert.Equal("conf_region", conf.getRegion(), "set config, empty env")
}

func TestGetEndpoint(t *testing.T) {
	assert := assert.New(t)

	os.Clearenv()
	defer os.Clearenv()

	conf := Config{}

	assert.Equal("", conf.getEndpoint(), "empty config, empty env")

	os.Setenv("AWS_ENDPOINT", "env_endpoint")
	assert.Equal("env_endpoint", conf.getEndpoint(), "empty config, set env")

	conf.Endpoint = "conf_endpoint"
	assert.Equal("conf_endpoint", conf.getEndpoint(), "set config, set env")

	os.Clearenv()
	assert.Equal("conf_endpoint", conf.getEndpoint(), "set config, empty env")
}

func TestEnvRegion(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		isSuccess bool
		envName   string
		region    string
	}{
		{true, "AWS_REGION", "foo"},
		{true, "AWS_REGION", "bar"},
		{false, "AWS_REGION1", "xxx"},
		{false, "AWS_REGION2", "xxx"},
		{false, "AWS_REGIO", "xxx"},
	}

	defer os.Clearenv()
	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)
		os.Clearenv()

		assert.Equal("", envRegion(), target)

		os.Setenv(tt.envName, tt.region)
		if !tt.isSuccess {
			assert.Equal("", envRegion(), target)
			return
		}

		assert.Equal(tt.region, envRegion(), target)
		os.Clearenv()
		assert.Equal("", envRegion(), target)
	}
}

func TestEnvEndpoint(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		isSuccess bool
		envName   string
		endpoint  string
	}{
		{true, "AWS_ENDPOINT", "foo"},
		{true, "AWS_ENDPOINT", "bar"},
		{false, "AWS_ENDPOINT1", "xxx"},
		{false, "AWS_ENDPOINT2", "xxx"},
		{false, "AWS_ENDPOIN", "xxx"},
	}

	defer os.Clearenv()
	for _, tt := range tests {
		target := fmt.Sprintf("%+v", tt)
		os.Clearenv()

		assert.Equal("", envEndpoint(), target)

		os.Setenv(tt.envName, tt.endpoint)
		if !tt.isSuccess {
			assert.Equal("", envEndpoint(), target)
			return
		}

		assert.Equal(tt.endpoint, envEndpoint(), target)
		os.Clearenv()
		assert.Equal("", envEndpoint(), target)
	}
}
