package storager

import (
	// _ "go.beyondstorage.io/services/azblob/v3"
	// _ "go.beyondstorage.io/services/fs/v4"
	// _ "go.beyondstorage.io/services/ftp"
	// _ "go.beyondstorage.io/services/gcs/v3"
	_ "go.beyondstorage.io/services/minio"
	_ "go.beyondstorage.io/services/s3/v3"

	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"
)

const (
	MinMultipartChunkSize = 5 * 1024 * 1024  // 5MB
	DefaultChunkSize      = 10 * 1024 * 1024 // 10MB
)

func New(connStr string) (types.Storager, error) {
	return services.NewStoragerFromString(connStr)
}
