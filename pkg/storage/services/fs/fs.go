package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/flowshot-io/x/pkg/storage/types"
)

type FileSystemBackend struct {
	Root        string
	pathMutexes sync.Map
}

func NewFileSystemBackend(root string) types.Storage {
	return &FileSystemBackend{Root: root}
}

func (fs *FileSystemBackend) ListWithContext(ctx context.Context, prefix string) (*[]types.Object, error) {
	mu := fs.getMutexForPath(prefix)
	mu.RLock()
	defer mu.RUnlock()

	var objects []types.Object
	fullPath := filepath.Join(fs.Root, prefix)

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		objectPath, _ := filepath.Rel(fs.Root, path)
		object := types.Object{
			Path:         objectPath,
			LastModified: info.ModTime(),
		}
		objects = append(objects, object)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &objects, nil
}

func (fs *FileSystemBackend) ReadWithContext(ctx context.Context, path string, writer io.Writer) (int64, error) {
	mu := fs.getMutexForPath(path)
	mu.RLock()
	defer mu.RUnlock()

	fullPath := filepath.Join(fs.Root, path)
	content, err := os.OpenFile(fullPath, os.O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer content.Close()

	return io.Copy(writer, content)
}

func (fs *FileSystemBackend) StatWithContext(ctx context.Context, path string) (*types.Object, error) {
	mu := fs.getMutexForPath(path)
	mu.RLock()
	defer mu.RUnlock()

	fullPath := filepath.Join(fs.Root, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	return &types.Object{
		Path:         path,
		LastModified: info.ModTime(),
	}, nil
}

func (fs *FileSystemBackend) WriteWithContext(ctx context.Context, path string, reader io.Reader, size int64) (int64, error) {
	mu := fs.getMutexForPath(path)
	mu.Lock()
	defer mu.Unlock()

	fullPath := filepath.Join(fs.Root, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return 0, err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return io.Copy(file, reader)
}

func (fs *FileSystemBackend) DeleteWithContext(ctx context.Context, path string) error {
	mu := fs.getMutexForPath(path)
	mu.Lock()
	defer mu.Unlock()

	fullPath := filepath.Join(fs.Root, path)
	return os.Remove(fullPath)
}

func (fs *FileSystemBackend) MoveWithContext(ctx context.Context, fromPath string, toPath string) error {
	srcMutex := fs.getMutexForPath(fromPath)
	dstMutex := fs.getMutexForPath(toPath)

	srcMutex.Lock()
	defer srcMutex.Unlock()
	dstMutex.Lock()
	defer dstMutex.Unlock()

	fromFullPath := filepath.Join(fs.Root, fromPath)
	toFullPath := filepath.Join(fs.Root, toPath)

	dir := filepath.Dir(toFullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(fromFullPath, toFullPath)
}

func (fs *FileSystemBackend) MoveToBucketWithContext(ctx context.Context, srcPath, dstPath, dstBucket string) error {
	srcMutex := fs.getMutexForPath(srcPath)
	dstMutex := fs.getMutexForPath(dstPath)

	srcMutex.Lock()
	defer srcMutex.Unlock()
	dstMutex.Lock()
	defer dstMutex.Unlock()

	srcFullPath := filepath.Join(fs.Root, srcPath)
	dstFullPath := filepath.Join(dstBucket, dstPath)

	dir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(srcFullPath, dstFullPath)
}

func (fs *FileSystemBackend) InitiateMultipartUploadWithContext(ctx context.Context, path string) (string, error) {
	return "", fmt.Errorf("fileSystemBackend does not support multipart uploads")
}

func (fs *FileSystemBackend) WriteMultipartWithContext(ctx context.Context, path, uploadID string, partNumber int64, reader io.ReadSeeker, size int64) (int64, *types.CompletedPart, error) {
	return size, nil, fmt.Errorf("fileSystemBackend does not support multipart uploads")
}

func (fs *FileSystemBackend) CompleteMultipartUploadWithContext(ctx context.Context, path, uploadID string, completedParts []*types.CompletedPart) error {
	return fmt.Errorf("fileSystemBackend does not support multipart uploads")
}

func (fs *FileSystemBackend) AbortMultipartUploadWithContext(ctx context.Context, path, uploadID string) error {
	return fmt.Errorf("fileSystemBackend does not support multipart uploads")
}

func (fs *FileSystemBackend) getMutexForPath(path string) *sync.RWMutex {
	mu, _ := fs.pathMutexes.LoadOrStore(path, &sync.RWMutex{})
	return mu.(*sync.RWMutex)
}
