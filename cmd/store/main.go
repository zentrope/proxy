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

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

//-----------------------------------------------------------------------------

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func generateArtifacts(artifactDir string, skus []*Sku) ([]*Sku, error) {
	updates := make([]*Sku, 0)
	for _, sku := range skus {
		dest := filepath.Join(artifactDir, sku.XRN+".zip")
		dest, err := filepath.Abs(dest)
		if err != nil {
			return nil, err
		}
		log.Printf("  - Bundling %v.", sku.Name)
		if err := zipit(sku.appdir, dest); err != nil {
			return nil, err
		}
		sku.zipfile = dest
		updates = append(updates, sku)
	}
	return updates, nil
}

//-----------------------------------------------------------------------------

type Sku struct {
	XRN         string `json:"xrn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Context     string `json:"context"`
	Download    string `json:"download_url"`
	//
	zipfile string
	appdir  string
}

func findSkus(src string) ([]*Sku, error) {

	sourceDir, err := filepath.Abs(src)
	if err != nil {
		return nil, err
	}

	sources, err := ioutil.ReadDir(sourceDir)

	skus := make([]*Sku, 0)

	for _, v := range sources {

		appName := v.Name()
		metadataFile := filepath.Join(sourceDir, appName, "metadata.js")

		bytes, err := ioutil.ReadFile(metadataFile)
		if err != nil {
			log.Printf("Unable to read metadata for app: %v (%v).", appName, err)
			continue
		}

		var sku Sku
		if err := json.Unmarshal(bytes, &sku); err != nil {
			log.Printf("Unable to parse metadata for app: %v (%v).",
				v.Name(), err)
			continue
		}
		sku.Context = appName
		sku.appdir = filepath.Join(sourceDir, v.Name())
		skus = append(skus, &sku)
	}

	return skus, nil
}

func toJson(skus []*Sku) (string, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(skus); err != nil {
		return "", err
	}
	return buf.String(), nil
}

//-----------------------------------------------------------------------------

func blockUntilShutdownThenDo(fn func()) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM,
		syscall.SIGKILL, syscall.SIGHUP)
	v := <-sigChan
	log.Printf("Signal: %v\n", v)
	fn()
}

type System struct {
	Skus      []*Sku
	sourceDir string
	deployDir string
}

func contexts(r *http.Request) []string {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return strings.Split(path, "/")
}

func (system *System) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("%v %v%v", r.Method, r.Host, r.URL.Path)

	context := contexts(r)

	switch context[0] {

	case "catalog":

		for i, sku := range system.Skus {
			path := "http://" + r.Host + "/download/" + sku.XRN + ".zip"
			system.Skus[i].Download = path
		}

		data, err := toJson(system.Skus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(data))
		return

	case "download":
		filename := filepath.Join(system.deployDir, context[1])
		f, err := os.Open(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		io.Copy(w, f)
		return

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h1>Not found.</h1>")
	}
}

func startupSequence() {

	sourceDir := "source"
	deployDir := "deploy"

	log.Println("Startup sequence.")

	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Fatalf("Can't find store source directory: %v", sourceDir)
	}

	if err := os.RemoveAll(deployDir); err != nil {
		log.Fatal(err)
	}

	if err := os.Mkdir(deployDir, 0755); err != nil {
		log.Fatal(err)
	}

	log.Println("- Finding apps.")

	skus, err := findSkus(sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("- Preparing app artifacts.")
	updates, err := generateArtifacts(deployDir, skus)
	if err != nil {
		log.Fatal(err)
	}

	render, err := toJson(updates)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(render)

	server := http.Server{
		Addr: ":60001",
		Handler: &System{
			Skus:      updates,
			deployDir: deployDir,
			sourceDir: sourceDir,
		},
	}

	log.Fatal(server.ListenAndServe())
}

func shutdownSequence() {
	log.Println("Shutdown sequence.")
}

func main() {
	log.Println("Welcome to Proxy App Store")

	go startupSequence()

	blockUntilShutdownThenDo(func() {
		shutdownSequence()
	})

	log.Println("System halt.")
}
