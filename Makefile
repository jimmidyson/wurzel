# Copyright 2015 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO   := GO15VENDOREXPERIMENT=1 go
pkgs  = $(shell $(GO) list ./... | grep -v /vendor/)
linters := --enable=vet --enable=deadcode --enable=golint --enable=varcheck --enable=structcheck --enable=aligncheck --enable=errcheck --enable=ineffassign --enable=interfacer --enable=goimports --disable=gocyclo --enable=gofmt --disable=gotype --disable=dupl

all: format build lint test bench

test:
	@echo ">> running tests"
	@$(GO) test -short -race -v $(pkgs) | tee /dev/tty | go-junit-report > $${CIRCLE_TEST_REPORTS:-.}/junit.xml

bench:
	@echo ">> running benchmarks"
	@for pkg in $(pkgs); do echo ">>> $$pkg"; $(GO) test -run=nothingplease -short -race -bench=. -benchmem $${pkg}; done

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

lint:
	@echo ">> linting code"
	@GOPATH=$(shell pwd)/vendor/:$${GOPATH} gometalinter --vendor --deadline=60s $(linters) ./...

build:
	@echo ">> building binaries"
	@./scripts/build.sh

docker:
	@docker build -t wurzel:$(shell git rev-parse --short HEAD) .

deps:
	@echo ">> installing dependencies"
	@go get -u github.com/alecthomas/gometalinter \
						 github.com/jstemmer/go-junit-report
	@gometalinter  --install --update --force

.PHONY: all format build test docker deps
