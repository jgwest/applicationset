package generators

import (
	"fmt"
	"testing"

	argoprojiov1alpha1 "github.com/argoproj-labs/applicationset/api/v1alpha1"
	"github.com/argoproj-labs/applicationset/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestListGenerateParams(t *testing.T) {

	type TestCase struct {
		name          string
		elements      []argoprojiov1alpha1.ListGeneratorElement
		expected      []map[string]string
		expectedError bool
	}

	testCases := []TestCase{
		{
			name:          "no-elements",
			elements:      []argoprojiov1alpha1.ListGeneratorElement{},
			expected:      []map[string]string{},
			expectedError: false,
		},
		{
			name: "single-element",
			elements: []argoprojiov1alpha1.ListGeneratorElement{
				{
					Cluster: "one",
					Url:     "https://1.2.3.4",
				},
			},
			expected: []map[string]string{
				{
					utils.ClusterListGeneratorKeyName: "one",
					utils.UrlGeneratorKeyName:         "https://1.2.3.4",
				},
			},
			expectedError: false,
		},
		{
			name: "multiple-elements",
			elements: []argoprojiov1alpha1.ListGeneratorElement{
				{
					Cluster: "one",
					Url:     "https://1.2.3.1",
				},
				{
					Cluster: "two",
					Url:     "https://1.2.3.2",
				},
				{
					Cluster: "three",
					Url:     "https://1.2.3.3",
				},
				{
					Cluster: "four",
					Url:     "https://1.2.3.4",
				},
				{
					Cluster: "five",
					Url:     "https://1.2.3.5",
				},
			},
			expected: []map[string]string{
				{
					utils.ClusterListGeneratorKeyName: "one",
					utils.UrlGeneratorKeyName:         "https://1.2.3.1",
				},
				{
					utils.ClusterListGeneratorKeyName: "two",
					utils.UrlGeneratorKeyName:         "https://1.2.3.2",
				},
				{
					utils.ClusterListGeneratorKeyName: "three",
					utils.UrlGeneratorKeyName:         "https://1.2.3.3",
				},
				{
					utils.ClusterListGeneratorKeyName: "four",
					utils.UrlGeneratorKeyName:         "https://1.2.3.4",
				},
				{
					utils.ClusterListGeneratorKeyName: "five",
					utils.UrlGeneratorKeyName:         "https://1.2.3.5",
				},
			},
			expectedError: false,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			listGenerator := NewListGenerator()

			generator := argoprojiov1alpha1.ApplicationSetGenerator{
				List: &argoprojiov1alpha1.ListGenerator{
					Elements: testCase.elements,
				},
				Git:      nil,
				Clusters: nil,
			}

			resultMap, err := listGenerator.GenerateParams(&generator)

			fmt.Println("result:", resultMap)

			assert.Equal(t, err != nil, testCase.expectedError)
			assert.ElementsMatch(t, resultMap, testCase.expected)

		})
	}
}

func TestListSanityChecks(t *testing.T) {
	listGenerator := NewListGenerator()

	result, err := listGenerator.GenerateParams(nil)
	assert.Empty(t, result)
	assert.Equal(t, err, EmptyAppSetGeneratorError)

	listGenerator = NewListGenerator()
	result, err = listGenerator.GenerateParams(&argoprojiov1alpha1.ApplicationSetGenerator{
		List: nil,
	})
	assert.Empty(t, result)
	assert.Equal(t, err, nil)
}
