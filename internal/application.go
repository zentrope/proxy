//
// Copyright (C) 2017 Keith Irwin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

func (a *Applications) AppMap() map[string]*InstalledApp {
	result := make(map[string]*InstalledApp, 0)
	for _, app := range a.InstalledApps {
		result[app.XRN] = app
	}

	return result
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
