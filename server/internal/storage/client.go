package storage

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	InternalEndpoint string
	UsePathStyle     bool
	Backet           string
}

type Client struct {
	bucket  string
	s3      *s3.Client
	presign *s3.PresignClient
}

func NewS3Client(c *Config) (*Client, error) {
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
		bucket:  c.Backet,
		s3:      client,
		presign: presignClient,
	}, nil
}
