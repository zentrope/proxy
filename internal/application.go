// Copyright 2017 Keith Irwin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

//-----------------------------------------------------------------------------

type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
}

type Application struct {
	Metadata Metadata `json:"metadata"`
	Icon     string   `json:"icon"`
	Context  string   `json:"context"`
}

type Applications struct {
	Dir  string         `json:"-"`
	Apps []*Application `json:"applications"`
}

func NewApplications(dir string) *Applications {

	return &Applications{
		Dir:  dir,
		Apps: nil,
	}
}

func (a *Applications) Reload() error {
	apps, err := findApps(a.Dir)
	if err != nil {
		return err
	}
	a.Apps = apps
	return nil
}

func (a *Applications) AsJSON() (string, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(a); err != nil {
		return "", err
	}
	return buf.String(), nil
}

//-----------------------------------------------------------------------------

func trimSvg(svg string) string {
	reg, err := regexp.Compile("[<][?].*[?][>]")
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(reg.ReplaceAllString(svg, ""))
}

func findApps(dir string) ([]*Application, error) {

	var icons = make(map[string]string, 0)
	var metadata = make(map[string]Metadata, 0)

	setIcon := func(parent, p string) error {
		bytes, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		icons[parent] = trimSvg(string(bytes))
		return nil
	}

	setMeta := func(parent, p string) error {
		bytes, err := ioutil.ReadFile(p)

		if err != nil {
			return err
		}

		var meta Metadata
		if err := json.Unmarshal(bytes, &meta); err != nil {
			return err
		}

		metadata[parent] = meta
		return nil
	}

	visit := func(p string, f os.FileInfo, err error) error {
		// Fragile because it's possible to find deeply nested svg and
		// metadata files using this general walker.

		if p == dir {
			return nil
		}

		parent := path.Base(path.Dir(p))

		if parent == path.Base(dir) {
			return nil
		}

		if path.Base(p) == "icon.svg" {
			if err := setIcon(parent, p); err != nil {
				return err
			}
		}

		if path.Base(p) == "metadata.js" {
			if err := setMeta(parent, p); err != nil {
				return err
			}
		}
		return nil
	}

	apps := make([]*Application, 0)

	err := filepath.Walk(dir, visit)
	if err != nil {
		return apps, err
	}

	for context, icon := range icons {
		apps = append(apps, &Application{
			Icon:     icon,
			Metadata: metadata[context],
			Context:  context,
		})
	}

	return apps, nil
}
