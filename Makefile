#
# Copyright (c) 2025 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#

export CGO_ENABLED=0

ifeq ($(shell go env GOOS),windows)
	BINARY=terraform-provider-osd-google.exe
	DESTINATION_PREFIX=$(APPDATA)/terraform.d/plugins
else
	BINARY=terraform-provider-osd-google
	DESTINATION_PREFIX=$(HOME)/.terraform.d/plugins
endif
OSDGOOGLE_LOCAL_DIR=$(DESTINATION_PREFIX)/terraform.local/local/osd-google

GO_ARCH=$(shell go env GOARCH)
TARGET_ARCH=$(shell go env GOOS)_${GO_ARCH}

import_path:=github.com/redhat/terraform-provider-osd-google
version=$(shell git describe --abbrev=0 2>/dev/null | sed 's/^v//' | sed 's/-prerelease\.[0-9]*//' || echo "0.0.1")
commit:=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

ldflags:=\
	-X $(import_path)/build.Version=$(version) \
	-X $(import_path)/build.Commit=$(commit) \
	$(NULL)

.PHONY: build
build:
	go build -ldflags="$(ldflags)" -o ${BINARY}

.PHONY: install
install: clean build
	platform=$$(terraform version -json | jq -r .platform); \
	extension=""; \
	if [[ "$${platform}" =~ ^windows_.*$$ ]]; then extension=".exe"; fi; \
	if [ -z "${version}" ]; then version="0.0.1"; fi; \
	dir="$(OSDGOOGLE_LOCAL_DIR)/$${version}/$(TARGET_ARCH)"; \
	file="terraform-provider-osd-google$${extension}"; \
	mkdir -p "$${dir}"; \
	mv ${BINARY} "$${dir}/$${file}"

.PHONY: subsystem-test
subsystem-test: install
	ginkgo run --succinct -ldflags="$(ldflags)" -r subsystem

.PHONY: acceptance-test
acceptance-test: install
	ginkgo run --succinct -ldflags="$(ldflags)" -r acceptance --label-filter "RequiresEnv"

.PHONY: unit-test
unit-test:
	go test -v ./provider/...

.PHONY: unit-test-coverage
unit-test-coverage:
	ginkgo run --succinct --cover --coverprofile coverage.out -ldflags="$(ldflags)" -r provider internal/...

.PHONY: test tests
test tests: unit-test subsystem-test

.PHONY: fmt_go
fmt_go:
	gofmt -s -l -w $$(find . -name '*.go' -not -path './references/*')

.PHONY: fmt_tf
fmt_tf:
	terraform fmt -recursive examples 2>/dev/null || true
	terraform fmt -recursive tests 2>/dev/null || true

.PHONY: fmt
fmt: fmt_go fmt_tf

.PHONY: clean
clean:
	rm -rf "$(OSDGOOGLE_LOCAL_DIR)"
	rm -f $(BINARY)

.PHONY: generate
generate:
	go generate ./...

.PHONY: tools
tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.13.2
	go install go.uber.org/mock/mockgen@v0.3.0
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.16.0

.PHONY: docs
docs:
	tfplugindocs generate --provider-name osdgoogle
