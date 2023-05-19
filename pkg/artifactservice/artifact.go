package artifactservice

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/flowshot-io/polystore/pkg/types"
	"github.com/flowshot-io/x/pkg/artifact"
)

// ArtifactServiceClient represents the methods required for artifact management.
type ArtifactServiceClient interface {
	UploadArtifact(ctx context.Context, artifact artifact.Artifact) error
	DownloadArtifact(ctx context.Context, artifactName string) (artifact.Artifact, error)
	DeleteArtifact(ctx context.Context, artifactName string) error
}

// Options holds the configuration for the artifact service.
type Options struct {
	Store      types.Storage
	WorkingDir string
}

// Client implements the ArtifactServiceClient interface.
type Client struct {
	store      types.Storage
	workingDir string
}

// New returns a new instance of an ArtifactServiceClient.
func New(opts Options) (ArtifactServiceClient, error) {
	if opts.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	if opts.WorkingDir == "" {
		opts.WorkingDir = "artifacts"
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

	if _, err := c.store.WriteWithContext(ctx, c.getWorkingPath(artifact.GetName()), bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
		return err
	}

	return nil
}

// DownloadArtifact downloads an artifact from storage.
func (c *Client) DownloadArtifact(ctx context.Context, artifactName string) (artifact.Artifact, error) {
	artifact := artifact.New(artifactName)
	path := c.getWorkingPath(artifact.GetName())

	if _, err := c.store.StatWithContext(ctx, path); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err := c.store.ReadWithContext(ctx, path, &buf); err != nil {
		return nil, err
	}

	if err := artifact.LoadFromReader(&buf); err != nil {
		return nil, err
	}

	return artifact, nil
}

// DeleteArtifact deletes an artifact from storage.
func (c *Client) DeleteArtifact(ctx context.Context, artifactName string) error {
	if err := c.store.DeleteWithContext(ctx, c.getWorkingPath(artifactName)); err != nil {
		return err
	}

	return nil
}

func (c *Client) getWorkingPath(artifactName string) string {
	return filepath.Join(c.workingDir, artifactName)
}
