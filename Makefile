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
	@echo ""
	@echo "  Per-example targets (each handles WIF config + cluster lifecycle):"
	@echo "  example.<name>          Apply WIF config then example (shorthand for .apply)"
	@echo "  example.<name>.init     Init terraform/wif_config and examples/<name>"
	@echo "  example.<name>.plan     Plan WIF config then example"
	@echo "  example.<name>.apply    Apply WIF config then example"
	@echo "  example.<name>.destroy  Destroy example then WIF config"
	@echo ""
	@echo "  Dev targets (install provider, clear lock, re-init, then run):"
	@echo "  dev.<name>              Apply with freshly installed provider"
	@echo "  dev.<name>.plan         Plan with freshly installed provider"
	@echo "  dev.<name>.apply        Apply with freshly installed provider"
	@echo "  dev.<name>.destroy      Destroy with freshly installed provider"
	@echo "  Examples: $(EXAMPLES)"
	@echo ""
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

# Example directories (discovered from examples/*)
EXAMPLES_DIRS := $(wildcard examples/*/main.tf)
EXAMPLES := $(sort $(patsubst examples/%/main.tf,%,$(EXAMPLES_DIRS)))

WIF_CONFIG_DIR := terraform/wif_config

# Inferred defaults (override with: make example.cluster.apply GCP_PROJECT_ID=my-proj CLUSTER_NAME=my-cluster)
GCP_PROJECT_ID ?= $(shell gcloud config get-value project 2>/dev/null)
CLUSTER_NAME ?= $(shell echo "$${USER:-$$(whoami)}" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/-\+/-/g' | sed 's/^-//;s/-$$//')
# Lazily evaluated so gcloud/whoami don't run during Makefile parse
TF_VARS = -var="gcp_project_id=$(GCP_PROJECT_ID)" -var="cluster_name=$(CLUSTER_NAME)"
EXTRA_TF_VARS ?=

# Per-example targets: each handles the full WIF + cluster lifecycle.
# Usage: make example.cluster.apply  make example.cluster_psc.plan
# Add extra vars: make example.cluster_shared_vpc.apply EXTRA_TF_VARS="-var=vpc_name=my-vpc"
# EXTRA_TF_VARS are only passed to the example dir, not the WIF config.
define example-targets
.PHONY: example.$(1) example.$(1).init example.$(1).plan example.$(1).apply example.$(1).destroy
example.$(1): example.$(1).apply

example.$(1).init:
	@cd $(WIF_CONFIG_DIR) && terraform init -upgrade
	@cd examples/$(1) && terraform init -upgrade

example.$(1).plan: example.$(1).init
	@[ -n "$$(GCP_PROJECT_ID)" ] || (echo "Error: GCP project not set. Run: gcloud config set project YOUR_PROJECT"; exit 1)
	@echo "Planning WIF config (project=$$(GCP_PROJECT_ID), cluster_name=$$(CLUSTER_NAME))..."
	@cd $(WIF_CONFIG_DIR) && terraform plan $$(TF_VARS)
	@echo ""
	@echo "Planning examples/$(1)..."
	@cd examples/$(1) && terraform plan $$(TF_VARS) $$(EXTRA_TF_VARS)

example.$(1).apply: example.$(1).init
	@[ -n "$$(GCP_PROJECT_ID)" ] || (echo "Error: GCP project not set. Run: gcloud config set project YOUR_PROJECT"; exit 1)
	@echo "Step 1: Creating WIF config in OCM (project=$$(GCP_PROJECT_ID), cluster_name=$$(CLUSTER_NAME))..."
	@cd $(WIF_CONFIG_DIR) && terraform apply -auto-approve $$(TF_VARS)
	@echo ""
	@echo "Step 2: Applying examples/$(1)..."
	@cd examples/$(1) && terraform apply -auto-approve $$(TF_VARS) $$(EXTRA_TF_VARS)

example.$(1).destroy: example.$(1).init
	@[ -n "$$(GCP_PROJECT_ID)" ] || (echo "Error: GCP project not set. Run: gcloud config set project YOUR_PROJECT"; exit 1)
	@echo "Step 1: Destroying examples/$(1)..."
	@cd examples/$(1) && terraform destroy -auto-approve $$(TF_VARS) $$(EXTRA_TF_VARS)
	@echo ""
	@echo "Step 2: Destroying WIF config..."
	@cd $(WIF_CONFIG_DIR) && terraform destroy -auto-approve $$(TF_VARS)
endef
$(foreach ex,$(EXAMPLES),$(eval $(call example-targets,$(ex))))

# Dev targets: install provider, clear lock, re-init, then run.
# Usage: make dev.cluster_with_vpc.apply  make dev.cluster_psc.plan
define dev-targets
.PHONY: dev.$(1) dev.$(1).init dev.$(1).plan dev.$(1).apply dev.$(1).destroy
dev.$(1): dev.$(1).apply

dev.$(1).init: install
	@rm -f $(WIF_CONFIG_DIR)/.terraform.lock.hcl examples/$(1)/.terraform.lock.hcl
	@cd $(WIF_CONFIG_DIR) && terraform init -upgrade
	@cd examples/$(1) && terraform init -upgrade

dev.$(1).plan: dev.$(1).init
	@[ -n "$$(GCP_PROJECT_ID)" ] || (echo "Error: GCP project not set. Run: gcloud config set project YOUR_PROJECT"; exit 1)
	@echo "Planning WIF config (dev)..."
	@cd $(WIF_CONFIG_DIR) && terraform plan $$(TF_VARS)
	@echo ""
	@echo "Planning examples/$(1) (dev)..."
	@cd examples/$(1) && terraform plan $$(TF_VARS) $$(EXTRA_TF_VARS)

dev.$(1).apply: dev.$(1).init
	@[ -n "$$(GCP_PROJECT_ID)" ] || (echo "Error: GCP project not set. Run: gcloud config set project YOUR_PROJECT"; exit 1)
	@echo "Step 1: Creating WIF config in OCM (dev)..."
	@cd $(WIF_CONFIG_DIR) && terraform apply -auto-approve $$(TF_VARS)
	@echo ""
	@echo "Step 2: Applying examples/$(1) (dev)..."
	@cd examples/$(1) && terraform apply -auto-approve $$(TF_VARS) $$(EXTRA_TF_VARS)

dev.$(1).destroy: dev.$(1).init
	@[ -n "$$(GCP_PROJECT_ID)" ] || (echo "Error: GCP project not set. Run: gcloud config set project YOUR_PROJECT"; exit 1)
	@echo "Step 1: Destroying examples/$(1) (dev)..."
	@cd examples/$(1) && terraform destroy -auto-approve $$(TF_VARS) $$(EXTRA_TF_VARS)
	@echo ""
	@echo "Step 2: Destroying WIF config (dev)..."
	@cd $(WIF_CONFIG_DIR) && terraform destroy -auto-approve $$(TF_VARS)
endef
$(foreach ex,$(EXAMPLES),$(eval $(call dev-targets,$(ex))))

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
	terraform fmt -recursive terraform 2>/dev/null || true
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
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.17.1
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
