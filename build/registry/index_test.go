package main

import (
	"github.com/falcosecurity/falcoctl/pkg/index"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_upsertIndex(t *testing.T) {
	tests := []struct {
		name              string
		registryPath      string
		ociArtifacts      map[string]string
		indexPath         string
		expectedIndexPath string
	}{
		{"missing", "testdata/registry.yaml", map[string]string{"falco": "ghcr.io/falcosecurity/rules/falco"}, "testdata/index1.yaml", "testdata/index_expected1.yaml"},
		{"already_present", "testdata/registry.yaml", map[string]string{"falco": "ghcr.io/falcosecurity/rules/falco"}, "testdata/index2.yaml", "testdata/index2.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := index.New(GHOrg)
			assert.NoError(t, i.Read(tt.indexPath))
			expectedIndex := index.New(GHOrg)
			assert.NoError(t, expectedIndex.Read(tt.expectedIndexPath))

			r, err := loadRegistryFromFile(tt.registryPath)
			assert.NoError(t, err)

			upsertIndex(r, tt.ociArtifacts, i)

			if !reflect.DeepEqual(i, expectedIndex) {
				t.Errorf("index() = %v, want %v", i, expectedIndex)
			}
		})
	}
}
