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
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type commandCode int

type commandResult struct {
	code    commandCode
	reason  string
	command command
}

type command interface {
	invoke(ctx *CommandProcessor) commandResult
}

type commandFunc func()

// CommandProcessor contains the state for running serialized commands.
type CommandProcessor struct {
	appDir    string
	database  *Database
	clienthub *ClientHub
	queue     chan commandFunc
}

const (
	commandOk    = commandCode(iota)
	commandError = iota
)

// NewCommandProcessor returns a processor for running serialized, side-effect commmands
func NewCommandProcessor(appDir string, database *Database, clients *ClientHub) *CommandProcessor {
	return &CommandProcessor{
		appDir:    appDir,
		database:  database,
		clienthub: clients,
		queue:     make(chan commandFunc),
	}
}

// Start the command processor.
func (cp *CommandProcessor) Start() {
	log.Printf("Starting command processor.")
	go cp.processJobs()
}

// Stop the command processor.
func (cp *CommandProcessor) Stop() {
	log.Printf("Stopping command processor.")
	close(cp.queue)
}

func (cp *CommandProcessor) invoke(clientID, commandTag, xrn string) {
	var cmd command
	switch commandTag {
	case "install":
		cmd = installCmd{xrn, cp.appDir}
	case "uninstall":
		cmd = uninstallCmd{xrn, cp.appDir}
	default:
		cmd = unknownCmd{commandTag}
	}

	cp.queue <- func() {
		result := cmd.invoke(cp)
		log.Printf("- %#v", result)
	}
}

//-----------------------------------------------------------------------------

func (cp *CommandProcessor) processJobs() {
	for job := range cp.queue {
		job()
	}
}

//-----------------------------------------------------------------------------

type installCmd struct {
	xrn    string
	appDir string
}

type uninstallCmd struct {
	xrn    string
	appDir string
}

type unknownCmd struct {
	name string
}

func (cmd installCmd) invoke(ctx *CommandProcessor) commandResult {

	sku, err := ctx.database.findSKU(cmd.xrn)
	if err != nil {
		return badResult(cmd, err.Error())
	}

	url := sku.DownloadURL

	publicDir, err := filepath.Abs(cmd.appDir)
	if err != nil {
		return badResult(cmd, err.Error())
	}

	target := filepath.Join(publicDir, filepath.Base(url))

	log.Printf("- Installing '%v'.", url)

	out, err := os.Create(target)
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return badResult(cmd, err.Error())
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return badResult(cmd, err.Error())
	}

	log.Printf("- Downloaded %v bytes.", n)

	if err := unzip(target, publicDir); err != nil {
		return badResult(cmd, err.Error())
	}

	if err := os.Remove(target); err != nil {
		return badResult(cmd, err.Error())
	}

	ctx.clienthub.notifyRefresh()
	return goodResult(cmd)
}

func (cmd uninstallCmd) invoke(ctx *CommandProcessor) commandResult {

	sku, err := ctx.database.findSKU(cmd.xrn)
	if err != nil {
		return badResult(cmd, err.Error())
	}

	publicDir, err := filepath.Abs(cmd.appDir)
	if err != nil {
		return badResult(cmd, err.Error())
	}

	contextDir := filepath.Join(publicDir, sku.Context)

	if err := os.RemoveAll(contextDir); err != nil {
		return badResult(cmd, err.Error())
	}

	ctx.clienthub.notifyRefresh()
	return goodResult(cmd)
}

func (cmd unknownCmd) invoke(ctx *CommandProcessor) commandResult {
	return badResult(cmd, fmt.Sprintf("Unknown command: '%v'.", cmd.name))
}

//-----------------------------------------------------------------------------

func goodResult(cmd command) commandResult {
	return commandResult{code: commandOk, reason: "Ok", command: cmd}
}

func badResult(cmd command, reason string) commandResult {
	return commandResult{code: commandError, reason: reason, command: cmd}
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
