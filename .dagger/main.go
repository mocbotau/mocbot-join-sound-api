package main

import (
	"context"

	"golang.org/x/sync/errgroup"

	"dagger/mocbot-join-sound-api/internal/dagger"
)

type MocbotJoinSoundApi struct {
	// Repository name
	// +private
	RepoName string
	// Source code directory
	// +private
	Source *dagger.Directory
	// +private
	InfisicalClientSecret *dagger.Secret
}

func New(
	repoName string,
	// Source code directory
	// +defaultPath="."
	source *dagger.Directory,
	// Infisical client secret
	infisicalClientSecret *dagger.Secret,
) *MocbotJoinSoundApi {
	return &MocbotJoinSoundApi{
		RepoName:              repoName,
		Source:                source,
		InfisicalClientSecret: infisicalClientSecret,
	}
}

// CI runs the complete CI pipeline
func (m *MocbotJoinSoundApi) CI(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return dag.GolangCi(m.Source).All(ctx)
	})

	g.Go(func() error {
		_, err := dag.Docker(m.Source, m.InfisicalClientSecret, m.RepoName).
			Build().
			GetContainer().
			Sync(ctx)

		return err
	})

	return g.Wait()
}

// BuildAndPush builds and pushes the Docker image to the container registry
func (m *MocbotJoinSoundApi) BuildAndPush(
	ctx context.Context,
	// +default="prod"
	env string,
) (string, error) {
	return dag.Docker(m.Source, m.InfisicalClientSecret, m.RepoName, dagger.DockerOpts{
		Environment: env,
	}).Build().Publish(ctx)
}
