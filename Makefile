# Copyright 2017 Keith Irwin. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

PACKAGE = github.com/zentrope/proxy

.PHONY: build run help build-client clean

.DEFAULT_GOAL := help

build-client: ## Build the client application
	cd client ; yarn build

build: ## Build the app.
	go build -o proxy

run: ## Run the app from source
	go run main.go

clean: ## Clean build artifacts
	-rm proxy
	-rm -rf client/build

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' | sort
