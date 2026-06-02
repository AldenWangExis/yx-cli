package app

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AldenWangExis/yx-cli/internal/gitx"
)

type CurrentRepositoryInput struct {
	ProfileName  string
	Domain       string
	Organization string
	Region       string
	WorkDir      string
	Remote       string
	Refresh      bool
}

type RepositoryID string

type RepositoryPath string

type GitRemoteName string

type CurrentRepository struct {
	ID           RepositoryID   `json:"id"`
	Name         string         `json:"name"`
	Path         RepositoryPath `json:"path"`
	Remote       GitRemoteName  `json:"remote"`
	RemoteURL    string         `json:"remoteUrl"`
	Domain       string         `json:"domain,omitempty"`
	Organization string         `json:"organization,omitempty"`
	Region       string         `json:"region,omitempty"`
	Source       string         `json:"source"`
}

type GitRemoteReader interface {
	ListRemotes(ctx context.Context, workDir string) ([]gitx.Remote, error)
}

type RepositoryIdentityCache interface {
	LookupRepositoryIdentity(key RepositoryIdentityCacheKey) (CurrentRepository, bool, error)
	StoreRepositoryIdentity(key RepositoryIdentityCacheKey, repo CurrentRepository) error
}

type RepositoryIdentityCacheKey struct {
	ProfileName  string
	Domain       string
	Organization string
	Region       string
	Path         string
}

func (k RepositoryIdentityCacheKey) StorageKey() string {
	if k.Domain == "" && k.Organization == "" && k.Region == "" {
		return k.Path
	}
	return strings.Join([]string{
		"v2",
		url.QueryEscape(k.Domain),
		url.QueryEscape(k.Region),
		url.QueryEscape(k.Organization),
		url.QueryEscape(k.Path),
	}, "|")
}

type RepositoryIdentityResolver struct {
	repositories RepositoryService
	git          GitRemoteReader
	cache        RepositoryIdentityCache
}

func NewRepositoryIdentityResolver(repositories RepositoryService, git GitRemoteReader, cache RepositoryIdentityCache) *RepositoryIdentityResolver {
	return &RepositoryIdentityResolver{repositories: repositories, git: git, cache: cache}
}

func (r *RepositoryIdentityResolver) CurrentRepository(ctx context.Context, input CurrentRepositoryInput) (CurrentRepository, error) {
	if input.Organization == "" {
		return CurrentRepository{}, fmt.Errorf("organization is required")
	}
	remote, err := r.currentCodeupRemote(ctx, input)
	if err != nil {
		return CurrentRepository{}, err
	}
	if remote.Organization != input.Organization {
		return CurrentRepository{}, fmt.Errorf("profile organization %q does not match remote organization %q", input.Organization, remote.Organization)
	}

	key := RepositoryIdentityCacheKey{
		ProfileName:  input.ProfileName,
		Domain:       strings.TrimRight(strings.TrimSpace(input.Domain), "/"),
		Organization: input.Organization,
		Region:       input.Region,
		Path:         remote.PathWithNamespace,
	}
	if !input.Refresh && r.cache != nil {
		if cached, ok, err := r.cache.LookupRepositoryIdentity(key); err == nil && ok && cached.ID != "" && string(cached.Path) == key.Path {
			cached.Remote = GitRemoteName(remote.RemoteName)
			cached.RemoteURL = remote.RemoteURL
			cached.Domain = key.Domain
			cached.Organization = key.Organization
			cached.Region = key.Region
			cached.Source = "cache"
			return cached, nil
		}
	}

	repos, err := r.repositories.ListRepositories(ctx)
	if err != nil {
		return CurrentRepository{}, err
	}
	matches := make([]RepositoryListItem, 0, 1)
	for _, repo := range repos {
		if repo.Path == key.Path {
			matches = append(matches, repo)
		}
	}
	if len(matches) == 0 {
		return CurrentRepository{}, fmt.Errorf("repository path %q was not found in Codeup API results", key.Path)
	}
	if len(matches) > 1 {
		return CurrentRepository{}, fmt.Errorf("repository path %q matched multiple Codeup API results", key.Path)
	}

	current := CurrentRepository{
		ID:           RepositoryID(matches[0].ID),
		Name:         matches[0].Name,
		Path:         RepositoryPath(matches[0].Path),
		Remote:       GitRemoteName(remote.RemoteName),
		RemoteURL:    remote.RemoteURL,
		Domain:       key.Domain,
		Organization: key.Organization,
		Region:       key.Region,
		Source:       "api",
	}
	if current.Name == "" {
		current.Name = remote.RepositoryName
	}
	if r.cache != nil {
		_ = r.cache.StoreRepositoryIdentity(key, current)
	}
	return current, nil
}

func (r *RepositoryIdentityResolver) currentCodeupRemote(ctx context.Context, input CurrentRepositoryInput) (gitx.CodeupRemote, error) {
	remotes, err := r.git.ListRemotes(ctx, input.WorkDir)
	if err != nil {
		return gitx.CodeupRemote{}, err
	}
	if input.Remote != "" {
		for _, remote := range remotes {
			if remote.Name != input.Remote {
				continue
			}
			return gitx.ParseCodeupRemoteURL(remote.Name, remote.URL)
		}
		return gitx.CodeupRemote{}, fmt.Errorf("git remote %q was not found", input.Remote)
	}

	codeupRemotes := []gitx.CodeupRemote{}
	for _, remote := range remotes {
		parsed, err := gitx.ParseCodeupRemoteURL(remote.Name, remote.URL)
		if err != nil {
			if strings.Contains(err.Error(), "credentials") {
				return gitx.CodeupRemote{}, err
			}
			continue
		}
		if parsed.RemoteName == "origin" {
			return parsed, nil
		}
		codeupRemotes = append(codeupRemotes, parsed)
	}
	if len(codeupRemotes) == 0 {
		return gitx.CodeupRemote{}, fmt.Errorf("no Codeup git remote found")
	}
	if len(codeupRemotes) > 1 {
		names := make([]string, 0, len(codeupRemotes))
		for _, remote := range codeupRemotes {
			names = append(names, remote.RemoteName)
		}
		return gitx.CodeupRemote{}, fmt.Errorf("multiple Codeup git remotes found (%s); specify --remote", strings.Join(names, ", "))
	}
	return codeupRemotes[0], nil
}
