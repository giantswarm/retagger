package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// missingCopies returns the tags a skopeo YAML file asks for that the
// `skopeo sync --digestfile` record at copiedPath has no line for, as
// "<name>:<tag>" , sorted.
//
// `skopeo sync --keep-going` logs a copy that failed and carries on, and exits
// zero when it has done so, so the record is the only place the failure shows:
// it holds one "<digest> docker://<registry>/<namespace>/<name>:<tag>" line per
// image that landed. The destination flattens the source repository to its last
// element, as imageBaseName does, so the comparison is on that name and the tag
// and leaves the registry out.
func missingCopies(filteredPath, copiedPath string) ([]string, error) {
	b, err := os.ReadFile(filteredPath)
	if err != nil {
		return nil, err
	}
	var wanted skopeoFile
	if err := yaml.Unmarshal(b, &wanted); err != nil {
		return nil, fmt.Errorf("parsing %q: %w", filteredPath, err)
	}

	f, err := os.Open(copiedPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	copied := map[string]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		// The last field is the reference, with or without a digest in front.
		reference := fields[len(fields)-1]
		reference = strings.TrimPrefix(reference, "docker://")
		copied[reference[strings.LastIndex(reference, "/")+1:]] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var missing []string
	seen := map[string]bool{}
	for _, registry := range wanted {
		for image, tags := range registry.Images {
			for _, tag := range tags {
				// A digest entry has no tag, so the record has no name:tag line to find.
				if strings.HasPrefix(tag, "sha256:") {
					logrus.Warnf("not verifying %s@%s: a digest has no line in the record", image, tag)
					continue
				}
				name := imageBaseName(image) + ":" + tag
				if seen[name] {
					logrus.Warnf("%s is asked for by more than one image: the copy flattens them to one name, so one copy covers all", name)
				}
				seen[name] = true
				if !copied[name] {
					missing = append(missing, name)
				}
			}
		}
	}
	sort.Strings(missing)
	return missing, nil
}

// commandVerifyCopied is invoked when `retagger verify-copied <filtered> <copied>`
// is called. It fails, naming each tag, when the images file asked for tags that
// the copy did not record.
func commandVerifyCopied(filteredPath, copiedPath string) {
	if filteredPath == "" || copiedPath == "" {
		logrus.Fatal("You need to specify both files: 'retagger verify-copied <filtered file> <copied record>'")
	}
	missing, err := missingCopies(filteredPath, copiedPath)
	if err != nil {
		logrus.Fatal(err)
	}
	if len(missing) == 0 {
		logrus.Infof("Every tag in %q was copied", filteredPath)
		return
	}
	for _, name := range missing {
		logrus.Errorf("not copied: %s", name)
	}
	logrus.Fatalf("%d tag(s) asked for by %q were not copied, see the skopeo log for the reason", len(missing), filteredPath)
}
