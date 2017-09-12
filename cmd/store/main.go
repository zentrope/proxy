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
	"time"
)

//-----------------------------------------------------------------------------

func blockUntilShutdownThenDo(fn func()) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM,
		syscall.SIGKILL, syscall.SIGHUP)
	v := <-sigChan
	log.Printf("Signal: %v\n", v)
	fn()
}

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

//-----------------------------------------------------------------------------

type System struct {
	Skus      []*Sku
	Cache     map[string]time.Time
	sourceDir string
	deployDir string
}

type Sku struct {
	XRN         string `json:"xrn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Context     string `json:"context"`
	Download    string `json:"download_url"`
	appdir      string
}

// All the top level directories in the source directory are
// considered the "root" of an application (or sku). The function
// returns a list of the imported metadata.js files for each
// application/sku.

func (system *System) LoadApps() error {

	sourceDir, err := filepath.Abs(system.sourceDir)
	if err != nil {
		return err
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

	system.Skus = skus
	return nil
}

func (system *System) CreateDownloads() error {
	for _, sku := range system.Skus {
		dest := filepath.Join(system.deployDir, sku.XRN+".zip")
		dest, err := filepath.Abs(dest)
		if err != nil {
			return err
		}
		log.Printf("  - Bundling %v.", sku.Name)
		if err := zipit(sku.appdir, dest); err != nil {
			return err
		}
	}
	return nil
}

func (system *System) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("%v %v%v", r.Method, r.Host, r.URL.Path)

	mkContexts := func(r *http.Request) []string {
		path := r.URL.Path
		path = strings.TrimPrefix(path, "/")
		path = strings.TrimSuffix(path, "/")
		return strings.Split(path, "/")
	}

	context := mkContexts(r)

	switch context[0] {

	case "catalog":

		for i, sku := range system.Skus {
			path := "http://" + r.Host + "/download/" + sku.XRN + ".zip"
			system.Skus[i].Download = path
		}

		w.Header().Set("content-type", "application/json")
		w.Write([]byte(system.String()))
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

func (system *System) String() string {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(system.Skus); err != nil {
		return fmt.Sprintf("<%v>", err)
	}
	return buf.String()
}

// On start up, the application removes the old zipped store bundles
// (skus) and recreates them based on the contents of a source
// directory.

func (system *System) Prepare() error {

	if _, err := os.Stat(system.sourceDir); os.IsNotExist(err) {
		log.Fatalf("Can't find store source directory: %v", system.sourceDir)
		return err
	}

	if err := os.RemoveAll(system.deployDir); err != nil {
		return err
	}

	if err := os.Mkdir(system.deployDir, 0755); err != nil {
		return err
	}

	return nil
}

func (system *System) Start() {

	log.Println("Startup sequence.")

	if err := system.Prepare(); err != nil {
		log.Fatalf("Unable to start: %v", err)
	}

	if err := system.LoadApps(); err != nil {
		log.Fatal(err)
	}

	if err := system.CreateDownloads(); err != nil {
		log.Fatal(err)
	}

	log.Println(system)

	server := http.Server{
		Addr:    ":60001",
		Handler: system,
	}

	log.Fatal(server.ListenAndServe())
}

func (system *System) Stop() {
	log.Println("Shutdown sequence.")
}

func main() {
	log.Println("Welcome to Proxy App Store")

	system := &System{sourceDir: "./source", deployDir: "./deploy"}

	go system.Start()

	blockUntilShutdownThenDo(func() {
		system.Stop()
	})

	log.Println("System halt.")
}
