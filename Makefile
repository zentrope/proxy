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

.PHONY: build run clean help init vendor vendor-list

.DEFAULT_GOAL := help

godep:
	@hash dep > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -v -u github.com/golang/dep/cmd/dep; \
	fi

vendor: godep ## Make sure vendor dependencies are present.
	dep ensure

vendor-list: ## List dependencies in vendor
	tree -C -d -L 2 vendor

build: vendor ## Build the app.
	go build -o backend cmd/backend/main.go
	go build -o proxy

run: ## Run the app from source
	go run main.go

clean: ## Clean build artifacts (if any)
	rm -f proxy
	rm -f backend
	rm -f cmd/backend/backend
	rm -rf cmd/store/deploy
	rm -rf public/beta
	rm -rf public/ten-forward
	rm -f  public/*.zip

dist-clean: clean ## Clean everything, including vendor.
	rm -rf vendor


init: vendor ## Initalize project (pull in vendors)

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' | sort
