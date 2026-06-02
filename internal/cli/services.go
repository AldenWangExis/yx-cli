package cli

import (
	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/gitx"
	"github.com/AldenWangExis/yx-cli/internal/safety"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/codeup"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/flow"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/platform"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao/projex"
)

type runtimeServices struct {
	runtime runtimeProfile
}

func (o Options) resolveRuntimeServices(ctx Context) (runtimeServices, error) {
	runtime, err := o.resolveRuntimeProfile(ctx)
	if err != nil {
		return runtimeServices{}, err
	}
	return runtimeServices{runtime: runtime}, nil
}

func (s runtimeServices) clientConfig() yunxiao.ClientConfig {
	return yunxiao.ClientConfig{
		BaseURL:        s.runtime.Profile.Domain,
		Token:          s.runtime.Token,
		OrganizationID: s.runtime.Profile.Organization,
		Region:         s.runtime.Profile.Region,
	}
}

func (s runtimeServices) safetyEnvironment() safety.Environment {
	return safety.Environment{ConfirmWrites: s.runtime.Profile.Safety.ConfirmWrites}
}

func (s runtimeServices) repoCurrentContext() repoCurrentRuntimeContext {
	return repoCurrentRuntimeContext{
		ProfileName:  s.runtime.Name,
		Domain:       yunxiao.NormalizeBaseURL(s.runtime.Profile.Domain),
		Organization: s.runtime.Profile.Organization,
		Region:       s.runtime.Profile.Region,
	}
}

func (s runtimeServices) repositoryAdapter() *codeup.RepositoryAdapter {
	return codeup.NewRepositoryAdapter(s.clientConfig())
}

func (s runtimeServices) changeRequestAdapter() *codeup.ChangeRequestAdapter {
	return codeup.NewChangeRequestAdapter(s.clientConfig())
}

func (s runtimeServices) projexAdapter() *projex.Adapter {
	return projex.NewAdapter(s.clientConfig())
}

func (s runtimeServices) flowAdapter() *flow.Adapter {
	return flow.NewAdapter(s.clientConfig())
}

func (s runtimeServices) platformAdapter() *platform.Adapter {
	return platform.NewAdapter(s.clientConfig())
}

func (s runtimeServices) repoUseCase() RepositoryUseCase {
	return app.NewRepoUseCase(s.repositoryAdapter(), gitx.NewRunner(), s.safetyEnvironment())
}

func (s runtimeServices) repoCurrentResolver() RepositoryCurrentResolver {
	return app.NewRepositoryIdentityResolver(
		s.repositoryAdapter(),
		gitx.NewRunner(),
		configRepositoryIdentityCache{store: s.runtime.Store},
	)
}

func (s runtimeServices) mergeRequestUseCase() MergeRequestUseCase {
	return app.NewMergeRequestUseCase(s.changeRequestAdapter(), s.safetyEnvironment())
}

func (s runtimeServices) workitemUseCase() WorkitemUseCase {
	adapter := s.projexAdapter()
	return app.NewWorkitemUseCase(adapter, adapter, s.runtime.Profile.RepoProjectMap, s.safetyEnvironment())
}

func (s runtimeServices) pipelineUseCase() PipelineUseCase {
	return app.NewPipelineUseCase(s.flowAdapter(), s.safetyEnvironment())
}

func (s runtimeServices) organizationUseCase() OrganizationUseCase {
	return app.NewOrganizationUseCase(s.platformAdapter())
}
