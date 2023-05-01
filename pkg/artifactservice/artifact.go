package artifactservice

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/flowshot-io/x/pkg/artifact"
	"github.com/flowshot-io/x/pkg/storager"
	"go.beyondstorage.io/v5/types"
)

type (
	ArtifactServiceClient interface {
		// UploadArtifact uploads an artifact to storage
		UploadArtifact(ctx context.Context, artifact *artifact.Artifact) error

		// GetArtifact gets an artifact from storage
		GetArtifact(ctx context.Context, artifactName string) (*artifact.Artifact, error)
	}

	Options struct {
		ConnectionString string
	}

	Client struct {
		store types.Storager
	}
)

func New(opts Options) (ArtifactServiceClient, error) {
	store, err := storager.New(opts.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to create store: %w", err)
	}

	return &Client{
		store: store,
	}, nil
}

func (c *Client) UploadArtifact(ctx context.Context, artifact *artifact.Artifact) error {
	var buf bytes.Buffer
	err := artifact.SaveToWriter(&buf)
	if err != nil {
		return err
	}

	_, err = c.store.WriteWithContext(ctx, artifact.Name, bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetArtifact(ctx context.Context, artifactName string) (*artifact.Artifact, error) {
	artifact := artifact.New(artifactName)

	_, err := c.store.StatWithContext(ctx, artifact.Name)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := io.Writer(&buf)

	_, err = c.store.ReadWithContext(ctx, artifact.Name, writer)
	if err != nil {
		return nil, err
	}

	err = artifact.LoadFromReader(&buf)
	if err != nil {
		return nil, err
	}

	return artifact, nil
}
