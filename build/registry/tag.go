package main

import (
	"fmt"
	"github.com/blang/semver"
	"regexp"
)

var (
	// see: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	// note: we have a capturing group for the plugin name prefix, so that we can use
	// it to specify the right make release target
	versionRegexp = regexp.MustCompile(`^([a-z]+[a-z0-9_\-]*)-((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?)$`)
)

type rulesfileNameSemver struct {
	Name   string
	Semver semver.Version
}

func (rn *rulesfileNameSemver) Version() string {
	return rn.Semver.String()
}

func parseGitTag(tag string) (*rulesfileNameSemver, error) {
	sm := versionRegexp.FindStringSubmatch(tag)
	if len(sm) == 0 {
		return nil, fmt.Errorf("tag %s could not be matched to a rulesfile name-version", tag)
	}

	rv := &rulesfileNameSemver{
		Name: sm[1],
	}

	sv, err := semver.Parse(sm[2])
	if err != nil {
		return nil, err
	}

	rv.Semver = sv

	return rv, nil
}

func isLatestSemver(newSemver semver.Version, existingSemvers []semver.Version) bool {
	for _, esv := range existingSemvers {
		if esv.GT(newSemver) {
			return false
		}
	}

	return true
}

func isLatestSemverForMinor(newSemver semver.Version, existingSemvers []semver.Version) bool {
	for _, esv := range existingSemvers {
		if esv.Minor == newSemver.Minor && esv.Major == newSemver.Major && esv.GT(newSemver) {
			return false
		}
	}

	return true
}

// ociTagsToUpdate returns the MAJOR.MINOR tag to update if any, the latest tag if any and the new tag to update
// in OCI registry given a new (already semver) tag and a list of existing tags in the OCI repo
func ociTagsToUpdate(newTag string, existingTags []string) []string {
	newSemver := semver.MustParse(newTag)

	if len(newSemver.Pre) > 0 {
		// pre-release version, do not update
		return nil
	}

	tagsToUpdate := []string{newSemver.String()}

	var existingSemvers []semver.Version
	for _, tag := range existingTags {
		if sv, err := semver.Parse(tag); err == nil {
			existingSemvers = append(existingSemvers, sv)
		}
	}

	if isLatestSemverForMinor(newSemver, existingSemvers) {
		tagsToUpdate = append(tagsToUpdate, fmt.Sprintf("%d.%d", newSemver.Major, newSemver.Minor))
	}

	if isLatestSemver(newSemver, existingSemvers) {
		tagsToUpdate = append(tagsToUpdate, "latest")
	}

	return tagsToUpdate
}
