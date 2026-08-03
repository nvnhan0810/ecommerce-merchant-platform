package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var (
	ErrObjectStoreDisabled = errors.New("object storage is not configured")
	ErrEmptyObjectKey      = errors.New("object key is required")
)

type Config struct {
	Endpoint     string
	Region       string
	AccessKey    string
	SecretKey    string
	Bucket       string
	UsePathStyle bool
}

type ObjectStore interface {
	Enabled() bool
	Upload(ctx context.Context, key string, body io.Reader, contentType string, size int64) error
	Download(ctx context.Context, key string) (body io.ReadCloser, contentType string, err error)
	Delete(ctx context.Context, key string) error
	NewProductImageKey(merchantID, productID, filename string) string
}

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3Store(cfg Config) (*S3Store, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" || strings.TrimSpace(cfg.Bucket) == "" {
		return nil, ErrObjectStoreDisabled
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, ErrObjectStoreDisabled
	}
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	client := s3.New(s3.Options{
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: cfg.UsePathStyle,
	})
	return &S3Store{client: client, bucket: cfg.Bucket}, nil
}

func (s *S3Store) Enabled() bool { return s != nil && s.client != nil && s.bucket != "" }

func (s *S3Store) Upload(ctx context.Context, key string, body io.Reader, contentType string, size int64) error {
	if !s.Enabled() {
		return ErrObjectStoreDisabled
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrEmptyObjectKey
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}
	_, err := s.client.PutObject(ctx, input)
	return err
}

func (s *S3Store) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	if !s.Enabled() {
		return nil, "", ErrObjectStoreDisabled
	}
	key = strings.TrimSpace(strings.TrimPrefix(key, "/"))
	if key == "" {
		return nil, "", ErrEmptyObjectKey
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		cancel()
		return nil, "", err
	}
	ct := "application/octet-stream"
	if out.ContentType != nil && *out.ContentType != "" {
		ct = *out.ContentType
	}
	return &cancelCloser{ReadCloser: out.Body, cancel: cancel}, ct, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	if !s.Enabled() {
		return ErrObjectStoreDisabled
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Store) NewProductImageKey(merchantID, productID, filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
	default:
		ext = ".jpg"
	}
	merchantID = strings.TrimSpace(merchantID)
	productID = strings.TrimSpace(productID)
	if merchantID == "" {
		merchantID = "unknown"
	}
	if productID == "" {
		productID = "unknown"
	}
	return fmt.Sprintf("shops/%s/products/%s/%s%s", merchantID, productID, uuid.NewString(), ext)
}

type cancelCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelCloser) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

// NopStore is used in tests / when S3 is unset.
type NopStore struct{}

func (NopStore) Enabled() bool { return false }
func (NopStore) Upload(context.Context, string, io.Reader, string, int64) error {
	return ErrObjectStoreDisabled
}
func (NopStore) Download(context.Context, string) (io.ReadCloser, string, error) {
	return nil, "", ErrObjectStoreDisabled
}
func (NopStore) Delete(context.Context, string) error { return ErrObjectStoreDisabled }
func (NopStore) NewProductImageKey(merchantID, productID, filename string) string {
	return fmt.Sprintf("shops/%s/products/%s/%s", merchantID, productID, filename)
}
