package main

import (
	"github.com/falcosecurity/falcoctl/pkg/index"
	"github.com/falcosecurity/falcoctl/pkg/oci"
	"path/filepath"
	"strings"
)

const (
	GHOrg = "falcosecurity"
)

func pluginRulesToIndexEntry(rf Rulesfile, registry, repo string) *index.Entry {
	return &index.Entry{
		Name:        rf.Name,
		Type:        string(oci.Rulesfile),
		Registry:    registry,
		Repository:  repo,
		Description: rf.Description,
		Home:        rf.URL,
		Keywords:    append(rf.Keywords, rf.Name),
		License:     rf.License,
		Maintainers: rf.Maintainers,
		Sources:     []string{rf.URL},
	}
}

func upsertIndex(r *Registry, ociArtifacts map[string]string, i *index.Index) {
	for _, rf := range r.Rulesfiles {
		// We only publish falcosecurity artifacts that have been uploaded to the repo.
		ref, ociRulesFound := ociArtifacts[rf.Name]

		// Build registry and repo starting from the reference.
		tokens := strings.Split(ref, "/")
		ociRegistry := tokens[0]
		ociRepo := filepath.Join(tokens[1:]...)
		if ociRulesFound {
			i.Upsert(pluginRulesToIndexEntry(rf, ociRegistry, ociRepo))
		}
	}
}

func upsertIndexFile(r *Registry, ociArtifacts map[string]string, indexPath string) error {
	i := index.New(GHOrg)

	if err := i.Read(indexPath); err != nil {
		return err
	}

	upsertIndex(r, ociArtifacts, i)

	return i.Write(indexPath)
}
