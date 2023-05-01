package artifactservice

import (
	"bytes"
	"context"
	"fmt"

	"github.com/flowshot-io/x/pkg/artifact"
	"github.com/flowshot-io/x/pkg/storager"
	"go.beyondstorage.io/v5/types"
)

// ArtifactServiceClient represents the methods required for artifact management.
type ArtifactServiceClient interface {
	UploadArtifact(ctx context.Context, artifact artifact.StorableArtifact) error
	GetArtifact(ctx context.Context, artifactName string) (artifact.StorableArtifact, error)
}

// Options holds the configuration for the artifact service.
type Options struct {
	ConnectionString string
	Store            types.Storager
}

// Client implements the ArtifactServiceClient interface.
type Client struct {
	store types.Storager
}

// New returns a new instance of an ArtifactServiceClient.
func New(opts Options) (ArtifactServiceClient, error) {
	if opts.Store == nil {
		store, err := storager.New(opts.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("unable to create store: %w", err)
		}

		opts.Store = store
	}

	return &Client{
		store: opts.Store,
	}, nil
}

// UploadArtifact uploads an artifact to storage.
func (c *Client) UploadArtifact(ctx context.Context, artifact artifact.StorableArtifact) error {
	var buf bytes.Buffer
	if err := artifact.SaveToWriter(&buf); err != nil {
		return err
	}

	if _, err := c.store.WriteWithContext(ctx, artifact.Name(), bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
		return err
	}

	return nil
}

// GetArtifact retrieves an artifact from storage.
func (c *Client) GetArtifact(ctx context.Context, artifactName string) (artifact.StorableArtifact, error) {
	artifact := artifact.New(artifactName)

	if _, err := c.store.StatWithContext(ctx, artifact.Name()); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := c.store.ReadWithContext(ctx, artifact.Name(), &buf); err != nil {
		return nil, err
	}

	if err := artifact.LoadFromReader(&buf); err != nil {
		return nil, err
	}

	return artifact, nil
}
