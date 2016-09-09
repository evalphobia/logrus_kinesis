package logrus_kinesis

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

const defaultRegion = "us-east-1"

// Config has AWS settings.
type Config struct {
	AccessKey string
	SecretKey string
	Region    string
	Endpoint  string
}

// AWSConfig creates *aws.Config object from the fields.
func (c Config) AWSConfig() *aws.Config {
	cred := c.awsCredentials()
	awsConf := &aws.Config{
		Credentials: cred,
		Region:      stringPtr(c.getRegion()),
	}

	ep := c.getEndpoint()
	if ep != "" {
		awsConf.Endpoint = &ep
	}

	return awsConf
}

func (c Config) awsCredentials() *credentials.Credentials {
	// from env
	cred := credentials.NewEnvCredentials()
	_, err := cred.Get()
	if err == nil {
		return cred
	}

	// from param
	cred = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
	_, err = cred.Get()
	if err == nil {
		return cred
	}

	// from local file
	return credentials.NewSharedCredentials("", "")
}

func (c Config) getRegion() string {
	if c.Region != "" {
		return c.Region
	}
	reg := envRegion()
	if reg != "" {
		return reg
	}
	return defaultRegion
}

func (c Config) getEndpoint() string {
	if c.Endpoint != "" {
		return c.Endpoint
	}
	ep := envEndpoint()
	if ep != "" {
		return ep
	}
	return ""
}

// envRegion get aws region from env params
func envRegion() string {
	return os.Getenv("AWS_REGION")
}

// envEndpoint get aws endpoint from env params
func envEndpoint() string {
	return os.Getenv("AWS_ENDPOINT")
}
