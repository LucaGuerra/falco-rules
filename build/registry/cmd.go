package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"os"
	"path/filepath"
	"strings"

	"github.com/falcosecurity/falcoctl/pkg/oci"
	"github.com/falcosecurity/falcoctl/pkg/oci/authn"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	RegistryTokenEnv = "REGISTRY_TOKEN"
	RegistryUserEnv  = "REGISTRY_USER"
	OCIRepoPrefixEnv = "OCI_REPO_PREFIX"
	RepoGithubEnv    = "GITHUB_REPO_URL"
)

func doCheck(fileName string) error {
	registry, err := loadRegistryFromFile(fileName)
	if err != nil {
		return err
	}
	return registry.Validate()
}

func doPushToOCI(registryFilename, gitTag string) error {
	var ociRepoPrefix, repoGit, user, token string
	var found bool

	if token, found = os.LookupEnv(RegistryTokenEnv); !found {
		return fmt.Errorf("environment variable with key %q not found, please set it before running this tool", RegistryTokenEnv)
	}

	if user, found = os.LookupEnv(RegistryUserEnv); !found {
		return fmt.Errorf("environment variable with key %q not found, please set it before running this tool", RegistryUserEnv)
	}

	if ociRepoPrefix, found = os.LookupEnv(OCIRepoPrefixEnv); !found {
		return fmt.Errorf("environment variable with key %q not found, please set it before running this tool", OCIRepoPrefixEnv)
	}

	if repoGit, found = os.LookupEnv(RepoGithubEnv); !found {
		return fmt.Errorf("environment variable with key %q not found, please set it before running this tool", RepoGithubEnv)
	}

	pt, err := parseGitTag(gitTag)
	if err != nil {
		return err
	}

	cred := auth.Credential{
		Username: user,
		Password: token,
	}

	client := authn.NewClient(cred)
	ociRepoRef := fmt.Sprintf("%s/%s", ociRepoPrefix, pt.Name)

	reg, err := loadRegistryFromFile(registryFilename)
	if err != nil {
		return fmt.Errorf("could not read registry from %s: %w", registryFilename, err)
	}

	rulesfileInfo := reg.RulesfileByName(pt.Name)
	if rulesfileInfo == nil {
		return fmt.Errorf("could not find rulesfile %s in registry", pt.Name)
	}

	existingTags, _ := oci.Tags(context.Background(), ociRepoRef, client)
	// note that the call above can generate an error if the repository does not exist yet, so we can proceed

	tagsToUpdate := ociTagsToUpdate(pt.Version(), existingTags)

	tmpDir, err := os.MkdirTemp("", "falco-artifacts-to-upload")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tgzFile := filepath.Join(tmpDir, filepath.Base(rulesfileInfo.Path)+".tar.gz")
	if err = tarGzSingleFile(tgzFile, rulesfileInfo.Path); err != nil {
		return fmt.Errorf("could not compress %s: %w", rulesfileInfo.Path, err)
	}
	defer os.RemoveAll(tgzFile)

	if err = pushCompressedRulesfile(client, tgzFile, ociRepoRef, repoGit, tagsToUpdate); err != nil {
		return fmt.Errorf("could not push %s to %s with source %s and tags %v: %w", tgzFile, ociRepoRef, repoGit, tagsToUpdate, err)
	}

	return nil
}

func rulesOciRepos(registryEntries *Registry, ociRepoPrefix string) (map[string]string, error) {
	ociClient := authn.NewClient(auth.EmptyCredential)
	ociEntries := make(map[string]string)

	for _, entry := range registryEntries.Rulesfiles {
		ref := ociRepoPrefix + "/" + entry.Name
		repo, err := remote.NewRepository(ref)
		if err != nil {
			return nil, fmt.Errorf("unable to create repo for ref %q: %w", ref, err)
		}
		repo.Client = ociClient

		_, _, err = repo.FetchReference(context.Background(), ref+":latest")
		if err != nil && (errors.Is(err, errdef.ErrNotFound) || strings.Contains(err.Error(), "requested access to the resource is denied")) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("unable to fetch reference for %q: %w", ref+":latest", err)
		}

		ociEntries[entry.Name] = ref
	}

	return ociEntries, nil
}

func doUpdateIndex(registryFile, indexFile string) error {
	var ociPrefix string
	var found bool

	if ociPrefix, found = os.LookupEnv(OCIRepoPrefixEnv); !found {
		return fmt.Errorf("environment variable with key %q not found, please set it before running this tool", OCIRepoPrefixEnv)
	}

	registryEntries, err := loadRegistryFromFile(registryFile)
	if err != nil {
		return err
	}
	ociEntries, err := rulesOciRepos(registryEntries, ociPrefix)
	if err != nil {
		return err
	}
	if err := registryEntries.Validate(); err != nil {
		return err
	}

	return upsertIndexFile(registryEntries, ociEntries, indexFile)
}

func main() {
	checkCmd := &cobra.Command{
		Use:                   "check <filename>",
		Short:                 "Verify the correctness of a registry YAML file",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			return doCheck(args[0])
		},
	}

	updateIndexCmd := &cobra.Command{
		Use:                   "update-index <registryFilename> <indexFilename>",
		Short:                 "Update an index file for artifacts distribution using registry data",
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			return doUpdateIndex(args[0], args[1])
		},
	}

	pushToOCI := &cobra.Command{
		Use:                   "push-to-oci <registryFilename> <gitTag>",
		Short:                 "Push the rulesfile identified by the tag to the OCI repo",
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			return doPushToOCI(args[0], args[1])
		},
	}

	rootCmd := &cobra.Command{
		Use:     "rules-registry",
		Version: "0.1.0",
	}
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(updateIndexCmd)
	rootCmd.AddCommand(pushToOCI)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
