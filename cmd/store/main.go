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

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
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

type system struct {
	skus      []*sku
	timestamp time.Time
	clock     *time.Ticker
	sourceDir string
	deployDir string
}

type sku struct {
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

func (system *system) loadApps() error {

	sourceDir, err := filepath.Abs(system.sourceDir)
	if err != nil {
		return err
	}

	sources, err := ioutil.ReadDir(sourceDir)

	skus := make([]*sku, 0)

	for _, v := range sources {

		appName := v.Name()
		metadataFile := filepath.Join(sourceDir, appName, "metadata.js")

		bytes, err := ioutil.ReadFile(metadataFile)
		if err != nil {
			log.Printf("Unable to read metadata for app: %v (%v).", appName, err)
			continue
		}

		var sku sku
		if err := json.Unmarshal(bytes, &sku); err != nil {
			log.Printf("Unable to parse metadata for app: %v (%v).",
				v.Name(), err)
			continue
		}
		sku.Context = appName
		sku.appdir = filepath.Join(sourceDir, v.Name())
		skus = append(skus, &sku)
	}

	system.timestamp = time.Now()
	system.skus = skus
	return nil
}

func (system *system) monitorApps() {

	// Blunt force: if anything changes in the app source directory,
	// repackage and re-deploy everything.

	recent := func(root string) (time.Time, error) {
		var t time.Time
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.ModTime().After(t) {
				t = info.ModTime()
			}
			return nil
		})

		return t, err
	}

	pass := func() error {
		sourceDir, err := filepath.Abs(system.sourceDir)
		if err != nil {
			return err
		}

		t, err := recent(sourceDir)
		if err != nil {
			return err
		}

		if t.After(system.timestamp) {
			log.Printf("Monitor: '%v' has been updated, and should be re-packaged.",
				system.sourceDir)
			system.loadApps()
			system.createDownloads()
		}

		return nil
	}

	c := system.clock.C
	for _ = range c {
		if err := pass(); err != nil {
			log.Printf("- ERROR: %v", err)
		}
	}
}

func (system *system) createDownload(sku *sku) error {
	dest := filepath.Join(system.deployDir, sku.XRN+".zip")
	dest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	log.Printf(" - Packaging %v.", sku.Name)
	return zipit(sku.appdir, dest)
}

func (system *system) createDownloads() error {
	for _, sku := range system.skus {
		if err := system.createDownload(sku); err != nil {
			return err
		}
	}
	return nil
}

func (system *system) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

		for i, sku := range system.skus {
			path := "http://" + r.Host + "/download/" + sku.XRN + ".zip"
			system.skus[i].Download = path
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

func (system *system) String() string {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(system.skus); err != nil {
		return fmt.Sprintf("<%v>", err)
	}
	return buf.String()
}

// On start up, the application removes the old zipped store bundles
// (skus) and recreates them based on the contents of a source
// directory.

func (system *system) prepare() error {

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

func (system *system) start() {

	log.Println("Starting system.")

	if err := system.prepare(); err != nil {
		log.Fatalf("Unable to start: %v", err)
	}

	if err := system.loadApps(); err != nil {
		log.Fatal(err)
	}

	if err := system.createDownloads(); err != nil {
		log.Fatal(err)
	}

	system.clock = time.NewTicker(11 * time.Second)
	go system.monitorApps()

	server := http.Server{
		Addr:    ":60001",
		Handler: system,
	}

	log.Fatal(server.ListenAndServe())
}

func (system *system) stop() {
	log.Println("Stopping system.")
	if system.clock != nil {
		system.clock.Stop()
	}
}

func newSystem(source, target string) *system {
	return &system{
		sourceDir: source,
		deployDir: target,
	}
}

func main() {
	log.Println("Welcome to Proxy App Store (port 60001).")

	var deployDir string
	var sourceDir string

	flag.StringVar(&deployDir, "d", "./deploy", "Directory for packaged apps.")
	flag.StringVar(&deployDir, "deploy", "./deploy", "Directory for packaged apps.")
	flag.StringVar(&sourceDir, "s", "./source", "Directory for application sources.")
	flag.StringVar(&sourceDir, "source", "./source", "Directory for application sources.")

	flag.Parse()

	log.Printf(" - Deploy pkgs: %v", deployDir)
	log.Printf(" - Source apps: %v", sourceDir)

	system := newSystem(sourceDir, deployDir)

	go system.start()
	blockUntilShutdownThenDo(system.stop)
	log.Println("System halt.")
}
