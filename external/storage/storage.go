package storage

import (
	"context"
	"io"
)

//go:generate go tool mockgen -source=storage.go -destination=../../test/mocks/storage/storage.go -package=mockstorage
type FileStorage interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error)
}
