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

.PHONY: build run run-backend run-store clean docker
.PHONY: help init vendor build-macos docker-build-macos

.DEFAULT_GOAL := help

#-----------------------------------------------------------------------------
# Build the app using Docker images such that you don't need an installed
# golang environment to `go run` the project and its dependent services.
#-----------------------------------------------------------------------------

DOCKCOMP = docker run --rm -v ${PWD}:/go/src/$(PACKAGE) -v ${PWD}:/go/bin golang:latest

## Fails if docker can't be found on the current system.

docker:
	@hash docker > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		echo "Unable to find docker command on path."; \
		exit 1; \
	fi

## For macOS

build-macos: init
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o proxy
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o store cmd/store/main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o backend cmd/backend/main.go

docker-build-macos: docker clean ## Use docker to compile app for macos.
	$(DOCKCOMP) bash -c "cd src/$(PACKAGE); make build-macos"

#-----------------------------------------------------------------------------
# Build goals
#-----------------------------------------------------------------------------

godep:
	@hash dep > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -v -u github.com/golang/dep/cmd/dep; \
	fi

vendor: godep ## Make sure vendor dependencies are present.
	dep ensure

build: vendor ## Build the app.
	go build -o backend cmd/backend/main.go
	go build -o store cmd/store/main.go
	go build -o proxy

clean: ## Clean build artifacts (if any).
	rm -f proxy
	rm -f backend
	rm -f store
	rm -f cmd/backend/backend
	rm -rf cmd/store/deploy
	rm -rf public/holodeck
	rm -rf public/ten-forward
	rm -f  public/*.zip
	rm -rf dep

dist-clean: clean ## Clean everything, including vendor.
	rm -rf vendor

init: vendor ## Initalize project (pull in vendors)

#-----------------------------------------------------------------------------
# Run goals: Run binaries if they exist, or from source if not.
#-----------------------------------------------------------------------------

run-backend: ## Run the backend service in the current terminal.
	@if [ -x ./backend ]; then \
		echo "** Running compiled backend api service." ; \
		./backend; \
	else \
		cd cmd/backend ; go run main.go; \
	fi

run-store: ## Run the app store service in the current terminal.
	@if [ -x ./store ]; then \
		echo "** Running compiled store service."; \
		./store -d /tmp/deploy -s cmd/store/source; \
	else \
		cd cmd/store ; go run main.go; \
	fi

run: ## Run proxy service in the current terminal.
	@if [ -x ./proxy ]; then \
	  echo "** Running compiled proxy service." && ./proxy; \
	else \
		$(MAKE) vendor ; go run main.go; \
	fi

#-----------------------------------------------------------------------------

help: ## Produce this list of goals
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' | \
		sort

#-----------------------------------------------------------------------------
