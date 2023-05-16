package artifactservice

import (
	"bytes"
	"context"
	"fmt"

	"github.com/flowshot-io/x/pkg/artifact"
	"github.com/flowshot-io/x/pkg/storage/types"
)

// ArtifactServiceClient represents the methods required for artifact management.
type ArtifactServiceClient interface {
	UploadArtifact(ctx context.Context, artifact artifact.Artifact) error
	GetArtifact(ctx context.Context, artifactName string) (artifact.Artifact, error)
}

// Options holds the configuration for the artifact service.
type Options struct {
	Store types.Storage
}

// Client implements the ArtifactServiceClient interface.
type Client struct {
	store types.Storage
}

// New returns a new instance of an ArtifactServiceClient.
func New(opts Options) (ArtifactServiceClient, error) {
	if opts.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	return &Client{
		store: opts.Store,
	}, nil
}

// UploadArtifact uploads an artifact to storage.
func (c *Client) UploadArtifact(ctx context.Context, artifact artifact.Artifact) error {
	var buf bytes.Buffer
	if err := artifact.SaveToWriter(&buf); err != nil {
		return err
	}

	if _, err := c.store.WriteWithContext(ctx, artifact.GetName(), bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
		return err
	}

	return nil
}

// GetArtifact retrieves an artifact from storage.
func (c *Client) GetArtifact(ctx context.Context, artifactName string) (artifact.Artifact, error) {
	artifact := artifact.New(artifactName)

	if _, err := c.store.StatWithContext(ctx, artifact.GetName()); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := c.store.ReadWithContext(ctx, artifact.GetName(), &buf); err != nil {
		return nil, err
	}

	if err := artifact.LoadFromReader(&buf); err != nil {
		return nil, err
	}

	return artifact, nil
}
