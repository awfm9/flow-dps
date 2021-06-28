// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:generate go run version.go

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"golang.org/x/mod/modfile"
)

const versionFileTemplate = `// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package configuration

const (
	RosettaVersion    = "{{ .RosettaVersion }}"
	NodeVersion       = "{{ .NodeVersion }}"
	MiddlewareVersion = "{{ .MiddlewareVersion }}"
)
`

const (
	defaultMiddlewareVersion = "0.0.0"
	rosettaVersion           = "1.4.10"

	pathToRepoRoot         = "../"
	pathToGoMod            = "../go.mod"
	rosettaVersionFilePath = "../rosetta/configuration/version.go"

	flowModPath = "github.com/onflow/flow-go"
)

func main() {

	fmt.Println("Using rosetta version", rosettaVersion)

	nodeVersion, err := NodeVersion()
	if err != nil {
		log.Fatalf("could not compute node version: %v", err)
	}

	fmt.Println("Found node version", nodeVersion)

	middlewareVersion, err := MiddlewareVersion()
	if err != nil {
		log.Fatalf("could not compute middleware version: %v", err)
	}

	fmt.Println("Found middleware version", middlewareVersion)

	tmpl := template.Must(template.New("version.go").Parse(versionFileTemplate))

	versionFile, err := os.Create(rosettaVersionFilePath)
	if err != nil {
		log.Fatalf("could not open version file: %v", err)
	}

	args := struct {
		RosettaVersion    string
		NodeVersion       string
		MiddlewareVersion string
	}{
		RosettaVersion:    rosettaVersion,
		NodeVersion:       nodeVersion,
		MiddlewareVersion: middlewareVersion,
	}

	err = tmpl.Execute(versionFile, args)
	if err != nil {
		log.Fatalf("could not execute template: %v", err)
	}
}

func NodeVersion() (string, error) {
	// Fetch Node version from the go.mod file.
	gomod, err := os.ReadFile(pathToGoMod)
	if err != nil {
		return "", fmt.Errorf("could not read go mod file: %w", err)
	}

	modfile, err := modfile.Parse("go.mod", gomod, nil)
	if err != nil {
		return "", fmt.Errorf("could not parse go mod file: %w", err)
	}

	for _, module := range modfile.Require {
		if module.Mod.Path == flowModPath {
			// Strip leading `v` from the tag if it exists.
			nodeVersion := strings.TrimPrefix(module.Mod.Version, "v")
			return nodeVersion, nil
		}
	}

	return "", fmt.Errorf("could not find github.com/onflow/flow-go dependency in go mod file")
}

func MiddlewareVersion() (string, error) {
	// Fetch middleware version by looking at the latest tag on the repository.
	repo, err := git.PlainOpen(pathToRepoRoot)
	if err != nil {
		return "", fmt.Errorf("could not open local git repository: %w", err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("could not find local git tags: %w", err)
	}

	// Fetch all tags and which commit they reference.
	tagsMap := make(map[plumbing.Hash]*plumbing.Reference)
	_ = tags.ForEach(func(t *plumbing.Reference) error {
		tagsMap[t.Hash()] = t
		return nil
	})

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("could not find local git HEAD: %w", err)
	}

	// Fetch the reference log.
	cIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return "", fmt.Errorf("could not read local git log: %w", err)
	}

	// Search for the latest tag on the current branch.
	var (
		tag   *plumbing.Reference
		count int
	)
	err = cIter.ForEach(func(c *object.Commit) error {
		if t, ok := tagsMap[c.Hash]; ok {
			tag = t
		}
		if tag != nil {
			return storer.ErrStop
		}
		count++
		return nil
	})
	if err != nil && !errors.Is(err, storer.ErrStop) {
		// Repository does not have any tags, return placeholder.
		return defaultMiddlewareVersion, nil
	}
	if tag == nil {
		// Repository does not have any relevant tags, return placeholder.
		return defaultMiddlewareVersion, nil
	}

	// Strip leading `v` from the tag if it exists.
	tagName := strings.TrimPrefix(tag.Name().Short(), "v")

	// If the current branch was just tagged, the version is precisely this tag.
	if count == 0 {
		return tagName, nil
	}

	// Otherwise, generate a version name from the latest tag name with HEAD's commit hash.
	t := fmt.Sprintf("%v-%v", tagName, head.Hash().String()[:8])
	return t, nil
}
