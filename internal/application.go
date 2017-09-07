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
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

//-----------------------------------------------------------------------------

type InstalledApp struct {
	XRN         string `json:"xrn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Icon        string `json:"icon"`
	Context     string `json:"context"`
}

type Applications struct {
	InstalledApps []*InstalledApp `json:"applications"`
	dir           string
}

func NewApplications(dir string) *Applications {

	return &Applications{
		dir:           dir,
		InstalledApps: nil,
	}
}

func (a *Applications) Reload() error {
	apps, err := findApps(a.dir)
	if err != nil {
		return err
	}
	a.InstalledApps = apps
	return nil
}

//-----------------------------------------------------------------------------

func trimSvg(svg string) string {
	reg, err := regexp.Compile("[<][?].*[?][>]")
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(reg.ReplaceAllString(svg, ""))
}

func findApps(dir string) ([]*InstalledApp, error) {

	contexts, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	installs := make([]*InstalledApp, 0)

	for _, contextDir := range contexts {

		if !contextDir.IsDir() {
			continue
		}

		appName := contextDir.Name()
		metadata := filepath.Join(dir, appName, "metadata.js")
		iconFile := filepath.Join(dir, appName, "icon.svg")

		bytes, err := ioutil.ReadFile(metadata)
		if err != nil {
			log.Printf("Unable to read metadata (%v) %v.", appName, err)
			return nil, err
		}

		var install InstalledApp
		if err := json.Unmarshal(bytes, &install); err != nil {
			log.Printf("Unable to parse metadata (%v) %v.", appName, err)
			return nil, err
		}

		iconBytes, err := ioutil.ReadFile(iconFile)
		if err != nil {
			log.Printf("Unable to read icon (%v) %v.", appName, err)
			return nil, err
		}

		install.Context = appName
		install.Icon = trimSvg(string(iconBytes))

		installs = append(installs, &install)
	}

	return installs, nil

}
