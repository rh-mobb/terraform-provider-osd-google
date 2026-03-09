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

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Terraform Provider for OSD on Google Cloud - Makefile targets"
	@echo ""
	@echo "  build             Build the provider binary"
	@echo "  install           Build and install to ~/.terraform.d/plugins"
	@echo "  dev-setup         Print dev_overrides config for ~/.terraformrc"
	@echo "  init-wif-example  Install provider and re-init examples/cluster_wif"
	@echo "  apply-wif-cluster Two-phase apply for examples/cluster_wif"
	@echo "  plan-wif-cluster  Plan for examples/cluster_wif"
	@echo "  destroy-wif-cluster Destroy examples/cluster_wif"
	@echo "  unit-test         Run unit tests"
	@echo "  subsystem-test    Run subsystem tests (requires install)"
	@echo "  acceptance-test   Run acceptance tests (requires OCM credentials)"
	@echo "  test, tests       Run unit + subsystem tests"
	@echo "  fmt               Format Go and Terraform code"
	@echo "  docs              Regenerate provider documentation"
	@echo "  clean             Remove build artifacts"
	@echo "  tools             Install dev tools (ginkgo, mockgen, tfplugindocs)"
	@echo "  generate          Run go generate"
	@echo "  references        Clone or update reference repos for AI agent context"
	@echo ""

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

import_path:=github.com/rh-mobb/terraform-provider-osd-google
version=$(shell git describe --abbrev=0 2>/dev/null | sed 's/^v//' | sed 's/-prerelease\.[0-9]*//' || echo "0.0.1")
commit:=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

ldflags:=\
	-X $(import_path)/build.Version=$(version) \
	-X $(import_path)/build.Commit=$(commit) \
	$(NULL)

PROVIDER_ADDRESS=registry.terraform.io/rh-mobb/osd-google

.PHONY: build
build:
	go build -ldflags="$(ldflags)" -o ${BINARY}

WIF_EXAMPLE_DIR := examples/cluster_wif

# Re-initialise the WIF example after installing the provider so the lock file
# reflects the new binary checksum.
.PHONY: init-wif-example
init-wif-example: install
	@rm -f $(WIF_EXAMPLE_DIR)/.terraform.lock.hcl
	@cd $(WIF_EXAMPLE_DIR) && terraform init -upgrade

.PHONY: apply-wif-cluster
apply-wif-cluster: init-wif-example
	@echo "Phase 1: Creating WIF config in OCM (OCM returns GCP blueprint)..."
	@cd $(WIF_EXAMPLE_DIR) && terraform apply -auto-approve -target=osdgoogle_wif_config.wif
	@echo ""
	@echo "Phase 2: Creating GCP resources (pool, service accounts, IAM) and cluster..."
	@cd $(WIF_EXAMPLE_DIR) && terraform apply -auto-approve

.PHONY: plan-wif-cluster
plan-wif-cluster: init-wif-example
	@cd $(WIF_EXAMPLE_DIR) && terraform plan

.PHONY: destroy-wif-cluster
destroy-wif-cluster: init-wif-example
	@echo "Destroying WIF cluster (using -refresh=false so for_each keys from wif_config state are known)..."
	@cd $(WIF_EXAMPLE_DIR) && terraform destroy -refresh=false -auto-approve

.PHONY: dev-setup
dev-setup: build
	@echo ""
	@echo "Provider binary built: $(CURDIR)/$(BINARY)"
	@echo ""
	@echo "Add the following to ~/.terraformrc to use the local build:"
	@echo ""
	@echo '  provider_installation {'
	@echo '    dev_overrides {'
	@echo '      "$(PROVIDER_ADDRESS)" = "$(CURDIR)"'
	@echo '    }'
	@echo '    direct {}'
	@echo '  }'
	@echo ""
	@echo "Then run terraform plan/apply in any example directory (no terraform init needed)."
	@echo ""

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
	ginkgo run --succinct -ldflags="$(ldflags)" -r provider internal/...

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
	terraform fmt -recursive modules 2>/dev/null || true

.PHONY: fmt
fmt: fmt_go fmt_tf

.PHONY: clean
clean:
	rm -rf "$(OSDGOOGLE_LOCAL_DIR)"
	rm -f $(BINARY)

.PHONY: generate
generate:
	go generate ./...

.PHONY: check-gen
check-gen: generate
	@git diff --exit-code || (echo "Generated files are not up to date. Run 'make generate' and commit." && exit 1)

.PHONY: tools
tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.13.2
	go install go.uber.org/mock/mockgen@v0.3.0
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.16.0

.PHONY: docs
docs:
	tfplugindocs generate --provider-name osdgoogle

# Reference repos for AI agent context (gitignored).
# Clone on first run; pull from main on subsequent runs.
define sync-ref
	@if [ -d "references/$(1)/.git" ]; then \
		echo "Updating references/$(1)..."; \
		git -C references/$(1) pull --ff-only; \
	else \
		echo "Cloning $(2) -> references/$(1)..."; \
		git clone --depth=1 $(2) references/$(1); \
	fi
endef

.PHONY: references
references:
	@mkdir -p references
	@echo "Syncing OCM OpenAPI spec..."
	@curl -sSL "https://raw.githubusercontent.com/openshift-online/ocm-sdk-go/main/openapi/clusters_mgmt/v1/openapi.json" \
	  -o references/OCM.json
	$(call sync-ref,ocm-sdk-go,https://github.com/openshift-online/ocm-sdk-go.git)
	$(call sync-ref,ocm-cli,https://github.com/openshift-online/ocm-cli.git)
	$(call sync-ref,terraform-provider-rhcs,https://github.com/terraform-redhat/terraform-provider-rhcs.git)
	$(call sync-ref,terraform-google-osd,https://github.com/rh-mobb/terraform-google-osd.git)
