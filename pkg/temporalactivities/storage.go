package temporalactivities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flowshot-io/x/pkg/storage/types"
)

type StorageActivities struct {
	storage types.Storage
}

func NewStorageActivities(storage types.Storage) *StorageActivities {
	return &StorageActivities{
		storage: storage,
	}
}

// MoveFile moves the specified file from the old path to the new path within the storage provider.
func (a *StorageActivities) MoveFile(ctx context.Context, fromPath string, toPath string) error {
	err := a.storage.MoveWithContext(ctx, fromPath, toPath)
	if err != nil {
		return err
	}

	return nil
}

// DeleteFile deletes the specified file from the storage provider.
func (a *StorageActivities) DeleteFile(ctx context.Context, path string) error {
	err := a.storage.DeleteWithContext(ctx, path)
	if err != nil {
		return err
	}

	return nil
}

// DownloadFile downloads the specified file from the storage provider to a local directory.
func (a *StorageActivities) DownloadFile(ctx context.Context, path string, destinationDir string) (string, error) {
	err := os.MkdirAll(destinationDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create destination directory: %v", err)
	}

	outputPath := filepath.Join(destinationDir, filepath.Base(path))
	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to file: %v", err)
	}
	defer file.Close()

	_, err = a.storage.ReadWithContext(ctx, path, file)
	if err != nil {
		return "", fmt.Errorf("failed to get object: %v", err)
	}

	return outputPath, nil
}

// UploadFile uploads the specified local file to the storage provider.
func (a *StorageActivities) UploadFile(ctx context.Context, path string, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stat: %v", err)
	}

	_, err = a.storage.WriteWithContext(ctx, destination, file, stat.Size())
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
