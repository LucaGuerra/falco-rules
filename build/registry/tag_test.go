package main

import (
	"github.com/blang/semver"
	"reflect"
	"sort"
	"testing"
)

// k8saudit-0.4.0-rc1
// k8saudit-eks-0.1.0
// dummy_c-0.3.1

func Test_parseTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    rulesfileNameSemver
		wantErr bool
	}{
		{"rc", "k8saudit-0.4.0-rc1", rulesfileNameSemver{
			Name:   "k8saudit",
			Semver: semver.MustParse("0.4.0-rc1"),
		}, false},
		{"eks", "k8saudit-extended-0.1.1", rulesfileNameSemver{
			Name:   "k8saudit-extended",
			Semver: semver.MustParse("0.1.1"),
		}, false},
		{"underscore", "dummy_c-eks-1.2.3-rc4", rulesfileNameSemver{
			Name:   "dummy_c-eks",
			Semver: semver.MustParse("1.2.3-rc4"),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitTag(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGitTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("parseGitTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ociTagsToUpdate(t *testing.T) {
	tests := []struct {
		name         string
		newTag       string
		existingTags []string
		want         []string
	}{
		{"latest", "0.3.2", []string{"0.1.1", "0.2.0", "0.3.1"}, []string{"0.3.2", "0.3", "latest"}},
		{"older", "0.1.1", []string{"0.1.2", "0.2.0", "0.3.1"}, []string{"0.1.1"}},
		{"latest_in_line", "0.1.3", []string{"0.1.2", "0.2.0", "0.3.1"}, []string{"0.1.3", "0.1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ociTagsToUpdate(tt.newTag, tt.existingTags)
			expected := tt.want

			sort.Strings(actual)
			sort.Strings(expected)

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("ociTagsToUpdate() = %v, want %v", actual, expected)
			}
		})
	}
}
