package main

import (
	"context"
	"fmt"
	"github.com/falcosecurity/falcoctl/pkg/oci"
	ocipusher "github.com/falcosecurity/falcoctl/pkg/oci/pusher"
	"k8s.io/klog/v2"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func pushCompressedRulesfile(ociClient *auth.Client, filePath, repoRef, repoGit string, tags []string) error {
	klog.Infof("Processing compressed rulesfile %q for repo %q and tags %s...", filePath, repoRef, tags)

	pusher := ocipusher.NewPusher(ociClient, false, nil)
	_, err := pusher.Push(context.Background(), oci.Rulesfile, repoRef,
		ocipusher.WithTags(tags...),
		ocipusher.WithFilepaths([]string{filePath}),
		ocipusher.WithAnnotationSource(repoGit))

	if err != nil {
		return fmt.Errorf("an error occurred while pushing: %w", err)
	}

	return nil
}
