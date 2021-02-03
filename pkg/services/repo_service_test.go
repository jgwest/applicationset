package services

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/reposerver/apiclient"
	"github.com/argoproj/argo-cd/util/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type ArgocdRepositoryMock struct {
	mock *mock.Mock
}

func (a ArgocdRepositoryMock) GetRepository(ctx context.Context, url string) (*v1alpha1.Repository, error) {
	args := a.mock.Called(ctx, url)

	return args.Get(0).(*v1alpha1.Repository), args.Error(1)

}

type repoServerClientMock struct {
	mock *mock.Mock
}

func (r repoServerClientMock) GenerateManifest(ctx context.Context, in *apiclient.ManifestRequest, opts ...grpc.CallOption) (*apiclient.ManifestResponse, error) {
	return nil, nil
}
func (r repoServerClientMock) ListApps(ctx context.Context, in *apiclient.ListAppsRequest, opts ...grpc.CallOption) (*apiclient.AppList, error) {
	args := r.mock.Called(ctx, in)

	return args.Get(0).(*apiclient.AppList), args.Error(1)
}
func (r repoServerClientMock) ListRefs(ctx context.Context, in *apiclient.ListRefsRequest, opts ...grpc.CallOption) (*apiclient.Refs, error) {
	args := r.mock.Called(ctx, in)

	return args.Get(0).(*apiclient.Refs), args.Error(1)
}
func (r repoServerClientMock) GetAppDetails(ctx context.Context, in *apiclient.RepoServerAppDetailsQuery, opts ...grpc.CallOption) (*apiclient.RepoAppDetailsResponse, error) {
	return nil, nil
}
func (r repoServerClientMock) GetRevisionMetadata(ctx context.Context, in *apiclient.RepoServerRevisionMetadataRequest, opts ...grpc.CallOption) (*v1alpha1.RevisionMetadata, error) {
	return nil, nil
}
func (r repoServerClientMock) GetHelmCharts(ctx context.Context, in *apiclient.HelmChartsRequest, opts ...grpc.CallOption) (*apiclient.HelmChartsResponse, error) {
	return nil, nil
}

type closer struct {
	// mock *mock.Mock
}

func (c closer) Close() error {
	return nil
}

type repoClientsetMock struct {
	mock *mock.Mock
}

func (r repoClientsetMock) NewRepoServerClient() (io.Closer, apiclient.RepoServerServiceClient, error) {
	args := r.mock.Called()

	return closer{}, args.Get(0).(apiclient.RepoServerServiceClient), args.Error(1)
}

func TestGetApps(t *testing.T) {

	for _, c := range []struct {
		name          string
		repoURL       string
		revision      string
		repoRes       *v1alpha1.Repository
		repoErr       error
		appRes        *apiclient.AppList
		appError      error
		expected      []string
		expectedError error
	}{
		{
			"Happy Flow",
			"repoURL",
			"revision",
			&v1alpha1.Repository{},
			nil,
			&apiclient.AppList{
				Apps: map[string]string{
					"app1": "",
					"app2": "",
				},
			},
			nil,
			[]string{"app1", "app2"},
			nil,
		},
		{
			"handles GetRepository error",
			"repoURL",
			"revision",
			&v1alpha1.Repository{},
			errors.New("error"),
			&apiclient.AppList{
				Apps: map[string]string{
					"app1": "",
					"app2": "",
				},
			},
			nil,
			[]string{},
			errors.New("Error in GetRepository: error"),
		},
		{
			"handles ListApps error",
			"repoURL",
			"revision",
			&v1alpha1.Repository{},
			nil,
			&apiclient.AppList{
				Apps: map[string]string{
					"app1": "",
					"app2": "",
				},
			},
			errors.New("error"),
			[]string{},
			errors.New("Error in ListApps: error"),
		},
	} {
		cc := c
		t.Run(cc.name, func(t *testing.T) {
			argocdRepositoryMock := ArgocdRepositoryMock{mock: &mock.Mock{}}
			repoServerClientMock := repoServerClientMock{mock: &mock.Mock{}}
			repoClientsetMock := repoClientsetMock{mock: &mock.Mock{}}

			argocdRepositoryMock.mock.On("GetRepository", mock.Anything, cc.repoURL).Return(cc.repoRes, cc.repoErr)

			repoServerClientMock.mock.On("ListApps", mock.Anything, &apiclient.ListAppsRequest{
				Repo:     cc.repoRes,
				Revision: cc.revision,
			}).Return(cc.appRes, cc.appError)

			repoClientsetMock.mock.On("NewRepoServerClient").Return(repoServerClientMock, nil)

			argocd := argoCDService{
				repositoriesDB: argocdRepositoryMock,
				repoClientset:  repoClientsetMock,
			}
			got, err := argocd.GetApps(context.TODO(), cc.repoURL, cc.revision)

			if cc.expectedError != nil {
				assert.EqualError(t, err, cc.expectedError.Error())
			} else {
				sort.Strings(got)
				sort.Strings(cc.expected)

				assert.Equal(t, got, cc.expected)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPaths(t *testing.T) {

	// Hardcode a specific commit, so that changes to argoproj/argocd-example-apps/ don't break our tests
	// "chore: downgrade kustomize guestbook image tag (#73)"
	commitID := "08f72e2a309beab929d9fd14626071b1a61a47f9"

	tests := []struct {
		name     string
		repoURL  string
		revision string
		pattern  string
		repoRes  *v1alpha1.Repository
		repoErr  error

		expectSubsetOfPaths []string
		doesNotContainPaths []string
		expectedError       error
	}{
		{
			name: "pull a specific revision of example apps and verify the list is expected",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:  "https://github.com/argoproj/argocd-example-apps/",
			revision: commitID,
			pattern:  "*",
			expectSubsetOfPaths: []string{
				"apps/Chart.yaml",
				"apps/templates/helm-guestbook.yaml",
				"apps/templates/helm-hooks.yaml",
				"apps/templates/kustomize-guestbook.yaml",
				"apps/templates/namespaces.yaml",
				"apps/templates/sync-waves.yaml",
				"apps/values.yaml",
				"blue-green/.helmignore",
				"blue-green/Chart.yaml",
				"blue-green/README.md",
				"blue-green/templates/NOTES.txt",
				"blue-green/templates/rollout.yaml",
				"blue-green/templates/services.yaml",
				"blue-green/values.yaml",
				"guestbook/guestbook-ui-deployment.yaml",
				"guestbook/guestbook-ui-svc.yaml",
				"kustomize-guestbook/guestbook-ui-deployment.yaml",
				"kustomize-guestbook/guestbook-ui-svc.yaml",
				"kustomize-guestbook/kustomization.yaml",
			},
		},
		{
			name: "pull an invalid revision, and confirm an error is returned",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:             "https://github.com/argoproj/argocd-example-apps/",
			revision:            "this-tag-does-not-exist",
			pattern:             "*",
			expectSubsetOfPaths: []string{},
			expectedError:       errors.New("Error during fetching commitSHA: Unable to resolve 'this-tag-does-not-exist' to a commit SHA"),
		},
		{
			name: "pull a specific revision of example apps, and use a ** pattern",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:  "https://github.com/argoproj/argocd-example-apps/",
			revision: commitID,
			pattern:  "**/*.yaml",
			expectSubsetOfPaths: []string{
				"apps/Chart.yaml",
				"apps/templates/helm-guestbook.yaml",
				"apps/templates/helm-hooks.yaml",
				"apps/templates/kustomize-guestbook.yaml",
				"apps/templates/namespaces.yaml",
				"apps/templates/sync-waves.yaml",
				"apps/values.yaml",
				"blue-green/templates/rollout.yaml",
				"blue-green/templates/services.yaml",
				"blue-green/values.yaml",
				"guestbook/guestbook-ui-deployment.yaml",
				"guestbook/guestbook-ui-svc.yaml",
				"kustomize-guestbook/guestbook-ui-deployment.yaml",
				"kustomize-guestbook/guestbook-ui-svc.yaml",
				"kustomize-guestbook/kustomization.yaml",
			},
			doesNotContainPaths: []string{
				"blue-green/.helmignore",
				"blue-green/README.md",
				"blue-green/templates/NOTES.txt",
			},
		},
	}

	for _, cc := range tests {

		// Get all the paths for a repository, and confirm that the expected subset of paths is found (or the expected error is returned)
		t.Run(cc.name, func(t *testing.T) {
			argocdRepositoryMock := ArgocdRepositoryMock{mock: &mock.Mock{}}

			argocdRepositoryMock.mock.On("GetRepository", mock.Anything, cc.repoURL).Return(cc.repoRes, cc.repoErr)

			argocd := argoCDService{
				repositoriesDB: argocdRepositoryMock,
			}

			getPathsRes, err := argocd.GetPaths(context.Background(), cc.repoURL, cc.revision, cc.pattern)

			if cc.expectedError == nil {

				assert.NoError(t, err)
				for _, path := range cc.expectSubsetOfPaths {
					assert.Contains(t, getPathsRes, path, "Unable to locate path: %s", path)
				}

				for _, shouldNotContain := range cc.doesNotContainPaths {
					assert.NotContains(t, getPathsRes, shouldNotContain, "GetPaths should not contain %s", shouldNotContain)
				}

			} else {
				assert.EqualError(t, err, cc.expectedError.Error())
			}
		})
	}

}

func TestGetFileContent(t *testing.T) {

	// Hardcode a specific commit, so that changes to argoproj/argocd-example-apps/ don't break our tests
	// "chore: downgrade kustomize guestbook image tag (#73)"
	commitID := "08f72e2a309beab929d9fd14626071b1a61a47f9"

	tests := []struct {
		name     string
		repoURL  string
		revision string
		path     string
		repoRes  *v1alpha1.Repository
		repoErr  error

		expectedFileContentsSubstring string
		expectedError                 error
	}{
		{
			name: "pull a specific path of a revision of example apps",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:                       "https://github.com/argoproj/argocd-example-apps/",
			revision:                      commitID,
			path:                          "/README.md",
			expectedFileContentsSubstring: "This repository contains example applications for demoing ArgoCD functionality",
		},
		{
			name: "pull another specific path of a revision of example apps",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:                       "https://github.com/argoproj/argocd-example-apps/",
			revision:                      commitID,
			path:                          "/helm-guestbook/Chart.yaml",
			expectedFileContentsSubstring: "A Helm chart for Kubernetes",
		},

		{
			name: "pull an invalid revison",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:       "https://github.com/argoproj/argocd-example-apps/",
			revision:      "this-tag-does-not-exist",
			path:          "/README.md",
			expectedError: errors.New("Error during fetching commitSHA: Unable to resolve 'this-tag-does-not-exist' to a commit SHA"),
		},
		{
			name: "pull an invalid file",
			repoRes: &v1alpha1.Repository{
				Insecure:              true,
				InsecureIgnoreHostKey: true,
				Repo:                  "https://github.com/argoproj/argocd-example-apps/",
			},
			repoURL:       "https://github.com/argoproj/argocd-example-apps/",
			revision:      commitID,
			path:          "/this-file-does-not-exist.md",
			expectedError: errors.New("open /tmp/https:__github.com_argoproj_argocd-example-apps_/this-file-does-not-exist.md: no such file or directory"),
		},
	}

	for _, cc := range tests {

		// Get the file contents of a specific path and revision, and confirm it has the expected value (or the expected error is returned.)
		t.Run(cc.name, func(t *testing.T) {

			argocdRepositoryMock := ArgocdRepositoryMock{mock: &mock.Mock{}}

			argocd := argoCDService{
				repositoriesDB: argocdRepositoryMock,
			}

			argocdRepositoryMock.mock.On("GetRepository", mock.Anything, cc.repoURL).Return(cc.repoRes, cc.repoErr)

			fileContentRes, err := argocd.GetFileContent(context.Background(), cc.repoURL, cc.revision, cc.path)

			if cc.expectedError == nil {

				// File contents should contain the expected string
				assert.True(t, strings.Contains(string(fileContentRes), cc.expectedFileContentsSubstring))
				assert.NoError(t, err)

			} else {
				assert.EqualError(t, err, cc.expectedError.Error())
			}

		})
	}

}
