package temporalactivities

import (
	"context"
	"os"
)

type (
	FSActivities struct{}
)

func NewFSActivities() *FSActivities {
	return &FSActivities{}
}

// MoveFile moves the specified file from the old path to the new path.
func MoveFile(ctx context.Context, fromPath string, toPath string) error {
	err := os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}

	return nil
}

// RemoveAll removes all files and directories at the specified path.
func (a *FSActivities) RemoveAll(ctx context.Context, path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}
