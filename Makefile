## Copyright 2017 Keith Irwin
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
##
## You may obtain a copy of the License at
##
##     http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
##
## See the License for the specific language governing permissions and
## limitations under the License.

PACKAGE = github.com/zentrope/proxy

.PHONY: build run clean help

.DEFAULT_GOAL := help

build: ## Build the app.
	go build -o backend cmd/backend/main.go
	go build -o proxy

run: ## Run the app from source
	go run main.go

clean: ## Clean build artifacts (if any)
	rm -f proxy
	rm -f backend
	rm -f cmd/backend/backend

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' | sort
