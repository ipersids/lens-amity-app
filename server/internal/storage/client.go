package storage

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	InternalEndpoint string
	PublicEndpoint   string
	UsePathStyle     bool
	Bucket           string
}

type Client struct {
	Bucket  string
	s3      *s3.Client
	Presign *s3.PresignClient
}

func NewS3Client(c Config) (*Client, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	// build aws.Config
	cfg := aws.Config{
		Region:       c.Region,
		BaseEndpoint: &c.InternalEndpoint,
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	}

	// build S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = c.UsePathStyle
	})

	if client == nil {
		return nil, errors.New("failed initialize S3 client from config")
	}

	// build S3 presign client
	presignClient := s3.NewPresignClient(client)

	if presignClient == nil {
		return nil, errors.New("failed initialize S3 presign client from config")
	}

	return &Client{
		Bucket:  c.Bucket,
		s3:      client,
		Presign: presignClient,
	}, nil
}

func (c Config) validate() error {
	if strings.TrimSpace(c.Region) == "" {
		return errors.New("storage: region is required")
	}
	if strings.TrimSpace(c.AccessKeyID) == "" {
		return errors.New("storage: access key id is required")
	}
	if strings.TrimSpace(c.SecretAccessKey) == "" {
		return errors.New("storage: secret access key is required")
	}
	if strings.TrimSpace(c.InternalEndpoint) == "" {
		return errors.New("storage: internal endpoint is required")
	}
	if strings.TrimSpace(c.PublicEndpoint) == "" {
		return errors.New("storage: public endpoint is required")
	}
	if strings.TrimSpace(c.Bucket) == "" {
		return errors.New("storage: bucket is required")
	}
	return nil
}
