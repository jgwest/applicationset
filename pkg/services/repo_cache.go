package services

import (
	"time"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/util/git"
	"github.com/pkg/errors"
)

type repoCache struct {
	cacheMap map[string]*repoCacheEntry
}

type repoCacheEntry struct {
	expirationTime time.Time
	gitClient      git.Client
	repoURL        string
}

func (cache *repoCache) GetOrGenerateCacheEntry(repo *v1alpha1.Repository) (*repoCacheEntry, error) {

	if repo == nil {
		return nil, errors.New("repository is nil")
	}

	repoURL := repo.Repo
	if repoURL == "" {
		return nil, errors.New("repository resolved to an empty string")
	}

	cacheEntry := cache.cacheMap[repoURL]
	if cacheEntry != nil { // If there is an entry in the cache for this repository...
		if time.Now().Before(cacheEntry.expirationTime) { // ...and if this repository's data has not yet expired...
			// ... then return it.
			return cacheEntry, nil
		}
	}

	// There is not an entry for this repoURL, OR it has expired.

	gitRepoClient, err := git.NewClient(repo.Repo, repo.GetGitCreds(), repo.IsInsecure(), repo.IsLFSEnabled())
	if err != nil {
		return nil, errors.Wrap(err, "Error in generating new client")
	}

	err = gitRepoClient.Init()
	if err != nil {
		return nil, errors.Wrap(err, "Error during initializing repo")
	}

	err = gitRepoClient.Fetch()
	if err != nil {
		return nil, errors.Wrap(err, "Error during fetching repo")
	}

	res := &repoCacheEntry{
		expirationTime: time.Now().Add(30 * time.Second),
		repoURL:        repo.Repo,
		gitClient:      gitRepoClient,
	}

	cache.cacheMap[repoURL] = res

	return res, nil
}
