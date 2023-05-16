package s3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	pathutil "path"
	"strings"

	"github.com/flowshot-io/x/pkg/storage/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/flowshot-io/x/pkg/storage/helpers"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Backend is a storage backend for S3
type S3Backend struct {
	Bucket     string
	Client     *s3.S3
	Downloader *s3manager.Downloader
	Prefix     string
	Uploader   *s3manager.Uploader
	SSE        string
}

// NewS3Backend creates a new instance of S3Backend
func NewS3Backend(bucket string, prefix string, region string, endpoint string, sse string) (types.Storage, error) {
	client := http.DefaultClient
	if os.Getenv("AWS_INSECURE_SKIP_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	sess, err := session.NewSession(&aws.Config{
		HTTPClient:       client,
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(strings.HasPrefix(endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(endpoint != ""),
	})
	if err != nil {
		return nil, err
	}

	service := s3.New(sess)
	b := &S3Backend{
		Bucket:     bucket,
		Client:     service,
		Downloader: s3manager.NewDownloaderWithClient(service),
		Prefix:     helpers.CleanPrefix(prefix),
		Uploader:   s3manager.NewUploaderWithClient(service),
		SSE:        sse,
	}
	return b, nil
}

// NewS3BackendWithCredentials creates a new instance of S3Backend with credentials
func NewS3BackendWithCredentials(bucket string, prefix string, region string, endpoint string, sse string, credentials *credentials.Credentials) (types.Storage, error) {
	client := http.DefaultClient
	if os.Getenv("AWS_INSECURE_SKIP_VERIFY") == "true" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}
	sess, err := session.NewSession(&aws.Config{
		HTTPClient:       client,
		Credentials:      credentials,
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(strings.HasPrefix(endpoint, "http://")),
		S3ForcePathStyle: aws.Bool(endpoint != ""),
	})
	if err != nil {
		return nil, err
	}

	service := s3.New(sess)
	b := &S3Backend{
		Bucket:     bucket,
		Client:     service,
		Downloader: s3manager.NewDownloaderWithClient(service),
		Prefix:     helpers.CleanPrefix(prefix),
		Uploader:   s3manager.NewUploaderWithClient(service),
		SSE:        sse,
	}
	return b, nil
}

// ListObjects lists all objects in a S3 bucket, at prefix
// Note: This function does not handle pagination and will return a maximum of 1000 objects.
// If there are more than 1000 objects with the specified prefix, consider using pagination to retrieve all objects.
func (b *S3Backend) ListWithContext(ctx context.Context, prefix string) (*[]types.Object, error) {
	var objects []types.Object
	prefix = pathutil.Join(b.Prefix, prefix)
	s3Input := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.Bucket),
		Prefix: aws.String(prefix),
	}

	s3Result, err := b.Client.ListObjectsV2WithContext(ctx, s3Input)
	if err != nil {
		return nil, err
	}

	for _, obj := range s3Result.Contents {
		path := helpers.RemovePrefixFromObjectPath(prefix, *obj.Key)
		if helpers.ObjectPathIsInvalid(path) {
			continue
		}
		object := types.Object{
			Path:         path,
			LastModified: *obj.LastModified,
		}
		objects = append(objects, object)
	}

	return &objects, nil
}

// ReadWithContext reads an object from S3 bucket, at given path
func (b *S3Backend) ReadWithContext(ctx context.Context, path string, writer io.Writer) (int64, error) {
	s3Input := &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
	}

	s3Result, err := b.Client.GetObjectWithContext(ctx, s3Input)
	if err != nil {
		return 0, err
	}
	defer s3Result.Body.Close()

	return io.Copy(writer, s3Result.Body)
}

// Stat returns information about an object in a S3 bucket, at given path
func (b *S3Backend) StatWithContext(ctx context.Context, path string) (*types.Object, error) {
	s3Input := &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
	}

	s3Result, err := b.Client.GetObjectWithContext(ctx, s3Input)
	if err != nil {
		return nil, err
	}

	return &types.Object{
		Path:         path,
		LastModified: *s3Result.LastModified,
	}, nil
}

// WriteObject uploads an object to S3, intelligently buffering large
// files into smaller chunks and sending them in parallel across multiple
// goroutines.
//
// Always use this method if your full file is already on disk or in memory.
func (b *S3Backend) WriteWithContext(ctx context.Context, path string, reader io.Reader, size int64) (int64, error) {
	s3Input := &s3manager.UploadInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
		Body:   reader,
	}

	if b.SSE != "" {
		s3Input.ServerSideEncryption = aws.String(b.SSE)
	}

	_, err := b.Uploader.UploadWithContext(ctx, s3Input)
	return size, err
}

// DeleteObject removes an object from a S3 bucket, at prefix
func (b *S3Backend) DeleteWithContext(ctx context.Context, path string) error {
	s3Input := &s3.DeleteObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
	}
	_, err := b.Client.DeleteObjectWithContext(ctx, s3Input)
	return err
}

// MoveObject moves an object from one path to another within a S3 bucket
func (b *S3Backend) MoveWithContext(ctx context.Context, fromPath string, toPath string) error {
	// First, copy the source file to the destination path
	copyInput := &s3.CopyObjectInput{
		Bucket:     aws.String(b.Bucket),
		CopySource: aws.String(url.PathEscape(pathutil.Join(b.Bucket, b.Prefix, fromPath))),
		Key:        aws.String(pathutil.Join(b.Prefix, toPath)),
	}
	if b.SSE != "" {
		copyInput.ServerSideEncryption = aws.String(b.SSE)
	}

	_, err := b.Client.CopyObjectWithContext(ctx, copyInput)
	if err != nil {
		return fmt.Errorf("failed to copy file from %s to %s: %w", fromPath, toPath, err)
	}

	// If the copy is successful, delete the source file
	err = b.DeleteWithContext(ctx, fromPath)
	if err != nil {
		return fmt.Errorf("failed to delete source file %s after copying: %w", fromPath, err)
	}

	return nil
}

// MoveToBucket moves an object from one S3 bucket to another
func (b *S3Backend) MoveToBucketWithContext(ctx context.Context, srcPath, dstPath, dstBucket string) error {
	// Step 1: Copy the object to the destination bucket
	copyInput := &s3.CopyObjectInput{
		Bucket:     aws.String(dstBucket),
		CopySource: aws.String(url.PathEscape(pathutil.Join(b.Bucket, b.Prefix, srcPath))),
		Key:        aws.String(pathutil.Join(b.Prefix, dstPath)),
	}

	_, err := b.Client.CopyObjectWithContext(ctx, copyInput)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	// Step 2: Delete the object from the source bucket
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, srcPath)),
	}

	_, err = b.Client.DeleteObjectWithContext(ctx, deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete object from source bucket: %w", err)
	}

	return nil
}

// InitiateMultipartUpload initiates a multipart upload.
//
// Use WriteWithContext over this method if the full file is already
// on the local machine as it will do the multipart upload for you.
func (b *S3Backend) InitiateMultipartUploadWithContext(ctx context.Context, path string) (string, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(pathutil.Join(b.Prefix, path)),
	}

	if b.SSE != "" {
		input.ServerSideEncryption = aws.String(b.SSE)
	}

	output, err := b.Client.CreateMultipartUploadWithContext(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	return *output.UploadId, nil
}

// WriteMultipartWithContext writes a part of a multipart upload
func (b *S3Backend) WriteMultipartWithContext(ctx context.Context, path, uploadID string, partNumber int64, reader io.ReadSeeker, size int64) (int64, *types.CompletedPart, error) {
	input := &s3.UploadPartInput{
		Bucket:        aws.String(b.Bucket),
		Key:           aws.String(pathutil.Join(b.Prefix, path)),
		UploadId:      aws.String(uploadID),
		PartNumber:    aws.Int64(partNumber),
		ContentLength: aws.Int64(size),
		Body:          reader,
	}

	output, err := b.Client.UploadPartWithContext(ctx, input)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to upload part: %w", err)
	}

	completedPart := &types.CompletedPart{
		ETag:       *output.ETag,
		Size:       size,
		PartNumber: partNumber,
	}

	return size, completedPart, nil
}

// CompleteMultipartUpload completes a multipart upload
func (b *S3Backend) CompleteMultipartUploadWithContext(ctx context.Context, path, uploadID string, completedParts []*types.CompletedPart) error {
	s3CompletedParts := make([]*s3.CompletedPart, len(completedParts))
	for i, part := range completedParts {
		s3CompletedParts[i] = &s3.CompletedPart{
			ETag:       aws.String(part.ETag),
			PartNumber: aws.Int64(part.PartNumber),
		}
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.Bucket),
		Key:      aws.String(pathutil.Join(b.Prefix, path)),
		UploadId: aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: s3CompletedParts,
		},
	}

	_, err := b.Client.CompleteMultipartUploadWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

// AbortMultipartUpload aborts the multipart upload
func (b *S3Backend) AbortMultipartUploadWithContext(ctx context.Context, path, uploadID string) error {
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(b.Bucket),
		Key:      aws.String(pathutil.Join(b.Prefix, path)),
		UploadId: aws.String(uploadID),
	}

	_, err := b.Client.AbortMultipartUploadWithContext(ctx, input)
	if err != nil {
		// AbortMultipartUpload is non-idempotent in s3, we need to omit `NoSuchUpload` error for compatibility.
		// AWS S3 is inconsistent in its behavior, it returns 204 No Content within 24 hours and returns 404 Not Found after 24hours.
		e := &s3types.NoSuchUpload{}
		if errors.As(err, &e) {
			err = nil
		}
	}
	if err != nil {
		return err
	}

	return nil
}
