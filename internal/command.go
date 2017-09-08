//-----------------------------------------------------------------------------
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
//-----------------------------------------------------------------------------

package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type CommandCode int

type CommandResult struct {
	Code   CommandCode
	Reason string
}

type CommandJob struct {
	command string
	oid     string
	result  chan CommandResult
}

type CommandProcessor struct {
	appDir   string
	appStore *AppStore
	queue    chan CommandJob
}

const (
	CommandOk    = CommandCode(iota)
	CommandError = iota
)

func NewCommandProcessor(appDir string, appStore *AppStore) *CommandProcessor {
	return &CommandProcessor{
		appDir:   appDir,
		appStore: appStore,
		queue:    make(chan CommandJob),
	}
}

func (cp *CommandProcessor) Start() {
	go cp.processJobs()
}

func (cp *CommandProcessor) Stop() {
	close(cp.queue)
}

func (cp *CommandProcessor) Invoke(clientId, command, oid string) chan CommandResult {
	out := make(chan CommandResult)
	go func() {
		cp.queue <- CommandJob{
			command: command,
			oid:     oid,
			result:  out,
		}
	}()
	return out
}

//-----------------------------------------------------------------------------

func (cp *CommandProcessor) processJobs() {
	for job := range cp.queue {
		defer close(job.result)
		switch job.command {

		case "install":
			job.result <- cp.install(job.oid)

		case "uninstall":
			job.result <- cp.uninstall(job.oid)

		default:
			job.result <- badResult(fmt.Sprintf("Unknown command: %v", job.command))
		}
	}
}

func (cp *CommandProcessor) uninstall(oid string) CommandResult {
	sku, err := cp.appStore.Find(oid)
	if err != nil {
		return badResult(err.Error())
	}

	publicDir, err := filepath.Abs(cp.appDir)
	if err != nil {
		return badResult(err.Error())
	}

	contextDir := filepath.Join(publicDir, sku.Context)

	if err := os.RemoveAll(contextDir); err != nil {
		return badResult(err.Error())
	}

	return goodResult()
}

func (cp *CommandProcessor) install(oid string) CommandResult {
	sku, err := cp.appStore.Find(oid)
	if err != nil {
		return badResult(err.Error())
	}

	url := sku.DownloadURL

	publicDir, err := filepath.Abs(cp.appDir)
	if err != nil {
		return badResult(err.Error())
	}

	target := filepath.Join(publicDir, filepath.Base(url))

	log.Printf("- Installing '%v'.", url)

	out, err := os.Create(target)
	defer out.Close()

	resp, err := http.Get(url)
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return badResult(err.Error())
	}

	log.Printf("- Downloaded %v bytes.", n)

	if err := unzip(target, publicDir); err != nil {
		return badResult(err.Error())
	}

	if err := os.Remove(target); err != nil {
		return badResult(err.Error())
	}

	return goodResult()
}

//-----------------------------------------------------------------------------

func goodResult() CommandResult {
	return CommandResult{Code: CommandOk, Reason: "Ok"}
}

func badResult(reason string) CommandResult {
	return CommandResult{Code: CommandError, Reason: reason}
}

func unzip(archive, target string) error {
	// http://blog.ralch.com/tutorial/golang-working-with-zip/
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

//-----------------------------------------------------------------------------
