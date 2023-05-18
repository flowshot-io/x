package temporalactivities

import (
	"context"

	"github.com/flowshot-io/x/pkg/artifact"
	"github.com/flowshot-io/x/pkg/artifactservice"
)

type ArtifactActivities struct {
	artifactClient artifactservice.ArtifactServiceClient
}

// NewArtifactActivities returns a new instance of an ArtifactActivities.
func NewArtifactActivities(artifactClient artifactservice.ArtifactServiceClient) *ArtifactActivities {
	return &ArtifactActivities{
		artifactClient: artifactClient,
	}
}

// PullArtifact downloads the specified artifact from the artifact service to a local directory.
func (a *ArtifactActivities) PullArtifact(ctx context.Context, artifactName string, destinationPath string) error {
	artifact, err := a.artifactClient.DownloadArtifact(ctx, artifactName)
	if err != nil {
		return err
	}

	err = artifact.ExtractToDirectory(destinationPath)
	if err != nil {
		return err
	}

	return nil
}

// PushArtifact creates an artifact from the specified files and uploads it to the artifact service.
func (a *ArtifactActivities) PushArtifact(ctx context.Context, artifactName string, files []string) error {
	artifact, err := artifact.NewWithPaths(artifactName, files)
	if err != nil {
		return err
	}

	err = a.artifactClient.UploadArtifact(ctx, artifact)
	if err != nil {
		return err
	}

	return nil
}
