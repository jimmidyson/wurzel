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

all: format build test

test:
	@echo ">> running tests"
	@$(GO) test -short -race $(pkgs)

bench:
	@echo ">> running benchmarks"
	@$(GO) test -bench=. -benchmem $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build:
	@echo ">> building binaries"
	@./scripts/build.sh

docker:
	@docker build -t wurzel:$(shell git rev-parse --short HEAD) .

ci-deps:
	@go get -u github.com/jstemmer/go-junit-report

ci: ci-deps
	@$(GO) test -short -race -v $(pkgs) | go-junit-report > report.xml

.PHONY: all format build test vet docker ci ci-deps
