package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/gitx"
)

func TestRepositoryIdentityResolverResolvesCurrentRepositoryFromAPIAndCachesIt(t *testing.T) {
	service := &fakeRepositoryService{
		list: []RepositoryListItem{
			{ID: "6562217", Name: "same-name", Path: "other-org/yx-cli"},
			{ID: "6925918", Name: "yx-cli", Path: "68086322e3a71588779435e0/yx-cli"},
		},
	}
	cache := newFakeRepositoryIdentityCache()
	resolver := NewRepositoryIdentityResolver(service, fakeGitRemoteReader{
		remotes: []gitx.Remote{{Name: "origin", URL: "git@codeup.aliyun.com:68086322e3a71588779435e0/yx-cli.git"}},
	}, cache)

	got, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "68086322e3a71588779435e0",
		WorkDir:      ".",
	})
	if err != nil {
		t.Fatalf("expected current repository, got: %v", err)
	}
	if got.ID != "6925918" || got.Name != "yx-cli" || got.Path != "68086322e3a71588779435e0/yx-cli" {
		t.Fatalf("unexpected current repository: %+v", got)
	}
	if got.Remote != "origin" || got.RemoteURL == "" || got.Source != "api" {
		t.Fatalf("unexpected remote/source: %+v", got)
	}
	if service.listCalls != 1 {
		t.Fatalf("expected one API list call, got %d", service.listCalls)
	}
	cacheKey := RepositoryIdentityCacheKey{
		ProfileName:  "default",
		Organization: "68086322e3a71588779435e0",
		Path:         "68086322e3a71588779435e0/yx-cli",
	}
	if cached, ok := cache.items[cacheKey.ProfileName+"|"+cacheKey.StorageKey()]; !ok || cached.ID != "6925918" {
		t.Fatalf("expected repository identity to be cached, got %+v", cache.items)
	}
}

func TestRepositoryIdentityResolverUsesExactPathNotRepositoryName(t *testing.T) {
	resolver := NewRepositoryIdentityResolver(&fakeRepositoryService{
		list: []RepositoryListItem{
			{ID: "wrong", Name: "yx-cli", Path: "other-org/yx-cli"},
		},
	}, fakeGitRemoteReader{
		remotes: []gitx.Remote{{Name: "origin", URL: "git@codeup.aliyun.com:68086322e3a71588779435e0/yx-cli.git"}},
	}, newFakeRepositoryIdentityCache())

	_, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "68086322e3a71588779435e0",
		WorkDir:      ".",
	})
	if err == nil || !strings.Contains(err.Error(), "68086322e3a71588779435e0/yx-cli") {
		t.Fatalf("expected exact path miss, got: %v", err)
	}
}

func TestRepositoryIdentityResolverRejectsProfileOrganizationMismatch(t *testing.T) {
	service := &fakeRepositoryService{}
	resolver := NewRepositoryIdentityResolver(service, fakeGitRemoteReader{
		remotes: []gitx.Remote{{Name: "origin", URL: "git@codeup.aliyun.com:remote-org/yx-cli.git"}},
	}, newFakeRepositoryIdentityCache())

	_, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "profile-org",
		WorkDir:      ".",
	})
	if err == nil || !strings.Contains(err.Error(), "profile organization") {
		t.Fatalf("expected organization mismatch, got: %v", err)
	}
	if service.listCalls != 0 {
		t.Fatalf("expected no API call on org mismatch, got %d", service.listCalls)
	}
}

func TestRepositoryIdentityResolverUsesCacheUnlessRefreshIsRequested(t *testing.T) {
	service := &fakeRepositoryService{
		list: []RepositoryListItem{{ID: "api-id", Name: "yx-cli", Path: "org/yx-cli"}},
	}
	cache := newFakeRepositoryIdentityCache()
	cacheKey := RepositoryIdentityCacheKey{ProfileName: "default", Organization: "org", Path: "org/yx-cli"}
	cache.items[cacheKey.ProfileName+"|"+cacheKey.StorageKey()] = CurrentRepository{ID: "cached-id", Name: "yx-cli", Path: "org/yx-cli"}
	resolver := NewRepositoryIdentityResolver(service, fakeGitRemoteReader{
		remotes: []gitx.Remote{{Name: "origin", URL: "git@codeup.aliyun.com:org/yx-cli.git"}},
	}, cache)

	cached, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "org",
		WorkDir:      ".",
	})
	if err != nil {
		t.Fatalf("expected cache hit, got: %v", err)
	}
	if cached.ID != "cached-id" || cached.Source != "cache" || service.listCalls != 0 {
		t.Fatalf("expected cache hit without API call, got repo=%+v calls=%d", cached, service.listCalls)
	}

	refreshed, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "org",
		WorkDir:      ".",
		Refresh:      true,
	})
	if err != nil {
		t.Fatalf("expected refresh to call API, got: %v", err)
	}
	if refreshed.ID != "api-id" || refreshed.Source != "api" || service.listCalls != 1 {
		t.Fatalf("expected API result on refresh, got repo=%+v calls=%d", refreshed, service.listCalls)
	}
}

func TestRepositoryIdentityResolverIgnoresCacheWriteFailure(t *testing.T) {
	cache := newFakeRepositoryIdentityCache()
	cache.storeErr = errors.New("disk full")
	resolver := NewRepositoryIdentityResolver(&fakeRepositoryService{
		list: []RepositoryListItem{{ID: "6925918", Name: "yx-cli", Path: "org/yx-cli"}},
	}, fakeGitRemoteReader{
		remotes: []gitx.Remote{{Name: "origin", URL: "git@codeup.aliyun.com:org/yx-cli.git"}},
	}, cache)

	got, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "org",
		WorkDir:      ".",
	})
	if err != nil {
		t.Fatalf("expected cache write failure not to block result, got: %v", err)
	}
	if got.ID != "6925918" || got.Source != "api" {
		t.Fatalf("unexpected current repository: %+v", got)
	}
}

func TestRepositoryIdentityResolverRequiresExplicitRemoteWhenMultipleCodeupRemotesExist(t *testing.T) {
	resolver := NewRepositoryIdentityResolver(&fakeRepositoryService{}, fakeGitRemoteReader{
		remotes: []gitx.Remote{
			{Name: "upstream", URL: "git@codeup.aliyun.com:org/a.git"},
			{Name: "fork", URL: "git@codeup.aliyun.com:org/b.git"},
		},
	}, newFakeRepositoryIdentityCache())

	_, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "org",
		WorkDir:      ".",
	})
	if err == nil || !strings.Contains(err.Error(), "--remote") {
		t.Fatalf("expected ambiguous remote error, got: %v", err)
	}
}

func TestRepositoryIdentityResolverHonorsExplicitRemote(t *testing.T) {
	resolver := NewRepositoryIdentityResolver(&fakeRepositoryService{
		list: []RepositoryListItem{{ID: "2", Name: "b", Path: "org/b"}},
	}, fakeGitRemoteReader{
		remotes: []gitx.Remote{
			{Name: "upstream", URL: "git@codeup.aliyun.com:org/a.git"},
			{Name: "fork", URL: "git@codeup.aliyun.com:org/b.git"},
		},
	}, newFakeRepositoryIdentityCache())

	got, err := resolver.CurrentRepository(context.Background(), CurrentRepositoryInput{
		ProfileName:  "default",
		Organization: "org",
		WorkDir:      ".",
		Remote:       "fork",
	})
	if err != nil {
		t.Fatalf("expected explicit remote to resolve, got: %v", err)
	}
	if got.ID != "2" || got.Remote != "fork" {
		t.Fatalf("unexpected explicit remote result: %+v", got)
	}
}

type fakeGitRemoteReader struct {
	remotes []gitx.Remote
	err     error
}

func (r fakeGitRemoteReader) ListRemotes(ctx context.Context, workDir string) ([]gitx.Remote, error) {
	return r.remotes, r.err
}

type fakeRepositoryIdentityCache struct {
	items    map[string]CurrentRepository
	storeErr error
}

func newFakeRepositoryIdentityCache() *fakeRepositoryIdentityCache {
	return &fakeRepositoryIdentityCache{items: map[string]CurrentRepository{}}
}

func (c *fakeRepositoryIdentityCache) LookupRepositoryIdentity(key RepositoryIdentityCacheKey) (CurrentRepository, bool, error) {
	item, ok := c.items[key.ProfileName+"|"+key.StorageKey()]
	return item, ok, nil
}

func (c *fakeRepositoryIdentityCache) StoreRepositoryIdentity(key RepositoryIdentityCacheKey, repo CurrentRepository) error {
	if c.storeErr != nil {
		return c.storeErr
	}
	c.items[key.ProfileName+"|"+key.StorageKey()] = repo
	return nil
}
