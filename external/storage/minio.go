package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/ishanwardhono/community-waste/pkg/config"
)

type Minio struct {
	client         *minio.Client
	bucket         string
	publicEndpoint string
	useSSL         bool
}

func NewMinio(cfg config.S3Config) (*Minio, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Minio{
		client:         client,
		bucket:         cfg.Bucket,
		publicEndpoint: cfg.PublicEndpoint,
		useSSL:         cfg.UseSSL,
	}, nil
}

func (m *Minio) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
}

func (m *Minio) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	scheme := "http"
	if m.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, m.publicEndpoint, m.bucket, key), nil
}
