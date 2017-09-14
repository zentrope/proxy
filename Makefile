##
## Copyright (C) 2017 Keith Irwin
##
## This program is free software: you can redistribute it and/or modify
## it under the terms of the GNU General Public License as published
## by the Free Software Foundation, either version 3 of the License,
## or (at your option) any later version.
##
## This program is distributed in the hope that it will be useful,
## but WITHOUT ANY WARRANTY; without even the implied warranty of
## MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
## GNU General Public License for more details.
##
## You should have received a copy of the GNU General Public License
## along with this program.  If not, see <http://www.gnu.org/licenses/>.

PACKAGE = github.com/zentrope/proxy

.PHONY: build run run-backend run-store clean
.PHONY: help init vendor vendor-list

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

run: vendor ## Run the app from source
	go run main.go

clean: ## Clean build artifacts (if any)
	rm -f proxy
	rm -f backend
	rm -f cmd/backend/backend
	rm -rf cmd/store/deploy
	rm -rf public/holodeck
	rm -rf public/ten-forward
	rm -f  public/*.zip

dist-clean: clean ## Clean everything, including vendor.
	rm -rf vendor

run-backend: ## Run the backend service in the current terminal
	cd cmd/backend ; go run main.go

run-store: ## Run the app store service in the current terminal.
	cd cmd/store ; go run main.go

init: vendor ## Initalize project (pull in vendors)

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' | sort
