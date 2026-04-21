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
	@echo "  wif.init                Init terraform/wif_config only"
	@echo "  wif.plan                Plan terraform/wif_config"
	@echo "  wif.apply               Apply terraform/wif_config (OCM WIF; run from repo root)"
	@echo "  wif.destroy             Destroy terraform/wif_config state"
	@echo "  (WIF / example / dev: variables from terraform.tfvars or TF_VAR_*; optional make TF_VARS=\"-var-file=...\")"
	@echo ""
	@echo "  Per-example targets (each handles WIF config + cluster lifecycle):"
	@echo "  example.<name>          Apply WIF config then example (shorthand for .apply)"
	@echo "  example.<name>.init     Init terraform/wif_config and examples/<name>"
	@echo "  example.<name>.plan     Plan WIF config then example"
	@echo "  example.<name>.apply    Apply WIF config then example"
	@echo "  example.<name>.destroy  Destroy example then WIF config"
	@echo "  example.<name>.login    oc login using api_url and admin credentials from terraform output"
	@echo "  example.cluster_private.ssh     Open IAP SSH session to the bastion VM"
	@echo "  example.cluster_private.tunnel  Forward cluster API to localhost:6443 via bastion"
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
	@echo "  lint              Run Go lint (golangci-lint) and Terraform validate"
	@echo "  docs              Regenerate provider documentation"
	@echo "  clean             Remove build artifacts"
	@echo "  tools             Install dev tools (ginkgo, mockgen, tfplugindocs, golangci-lint)"
	@echo "  generate          Run go generate"
	@echo "  references        Clone or update reference repos for AI agent context"
	@echo ""

SHELL := /bin/bash

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

# Module directories (discovered from modules/*)
MODULES := $(sort $(notdir $(wildcard modules/*/)))

WIF_CONFIG_DIR := terraform/wif_config

# WIF-only lifecycle (terraform/wif_config). Run from repository root; variables come from
# terraform/wif_config/terraform.tfvars or TF_VAR_* (same as example targets).
.PHONY: wif.init wif.plan wif.apply wif.destroy
wif.init:
	@cd $(WIF_CONFIG_DIR) && terraform init -upgrade

wif.plan: wif.init
	@cd $(WIF_CONFIG_DIR) && terraform plan $(TF_VARS)

wif.apply: wif.init
	@echo "Applying $(WIF_CONFIG_DIR)..."
	@cd $(WIF_CONFIG_DIR) && terraform apply -auto-approve $(TF_VARS)

wif.destroy: wif.init
	@cd $(WIF_CONFIG_DIR) && terraform destroy -auto-approve $(TF_VARS)

# Optional extra CLI args for terraform (e.g. TF_VARS="-var-file=custom.tfvars").
# Variables are otherwise supplied by terraform.tfvars and/or TF_VAR_* in the environment.
TF_VARS :=
EXTRA_TF_VARS ?=

# Per-example targets: each handles the full WIF + cluster lifecycle.
# Usage: make example.cluster.apply  make example.cluster_psc.plan
# Add extra vars: make example.cluster_shared_vpc.apply EXTRA_TF_VARS="-var=vpc_name=my-vpc"
# EXTRA_TF_VARS are only passed to the example dir, not the WIF config.
define example-targets
.PHONY: example.$(1) example.$(1).init example.$(1).plan example.$(1).apply example.$(1).destroy example.$(1).login
example.$(1): example.$(1).apply

example.$(1).init:
	@cd $(WIF_CONFIG_DIR) && terraform init -upgrade
	@cd examples/$(1) && terraform init -upgrade

example.$(1).plan: example.$(1).init
	@echo "Planning WIF config..."
	@cd $(WIF_CONFIG_DIR) && terraform plan $(TF_VARS)
	@echo ""
	@echo "Planning examples/$(1)..."
	@cd examples/$(1) && terraform plan $(TF_VARS) $(EXTRA_TF_VARS)

example.$(1).apply: example.$(1).init
	@echo "Step 1: Creating WIF config in OCM..."
	@cd $(WIF_CONFIG_DIR) && terraform apply -auto-approve $(TF_VARS)
	@echo ""
	@echo "Step 2: Applying examples/$(1)..."
	@cd examples/$(1) && terraform apply -auto-approve $(TF_VARS) $(EXTRA_TF_VARS)

example.$(1).destroy: example.$(1).init
	@echo "Step 1: Destroying examples/$(1)..."
	@cd examples/$(1) && terraform destroy -auto-approve $(TF_VARS) $(EXTRA_TF_VARS)
	@echo ""
	@echo "Step 2: Destroying WIF config..."
	@cd $(WIF_CONFIG_DIR) && terraform destroy -auto-approve $(TF_VARS)

# Use shell backticks for terraform output, not $(...): GNU Make expands $(...) in
# recipes as Make variables. Also avoid $$API — Make parses that as $A + PI (variable A).
example.$(1).login:
	@command -v oc >/dev/null 2>&1 || (echo "Error: OpenShift CLI (oc) not found on PATH."; exit 1)
	@cd examples/$(1) && \
		terraform output -raw api_url >/dev/null 2>&1 || { echo "Error: terraform output failed. Initialize and apply this example first (e.g. make example.$(1).apply)."; exit 1; }; \
		test -n "`terraform output -raw api_url`" || { echo "Error: api_url is empty in Terraform state."; exit 1; }; \
		oc login "`terraform output -raw api_url`" \
			-u "`terraform output -raw admin_username`" \
			-p "`terraform output -raw admin_password`"
endef
$(foreach ex,$(EXAMPLES),$(eval $(call example-targets,$(ex))))

# Private cluster: open an interactive IAP SSH session to the bastion VM.
# Requires: gcloud CLI, IAP API enabled, bastion created (example.cluster_private.apply).
.PHONY: example.cluster_private.ssh
example.cluster_private.ssh:
	@command -v gcloud >/dev/null 2>&1 || (echo "Error: gcloud CLI not found on PATH."; exit 1)
	@cd examples/cluster_private && \
		BASTION=`terraform output -raw bastion_name` && \
		ZONE=`terraform output -raw bastion_zone` && \
		PROJECT=`terraform output -raw gcp_project_id` && \
		echo "Connecting to $$BASTION ($$ZONE) via IAP..." && \
		gcloud compute ssh "$$BASTION" --project="$$PROJECT" --zone="$$ZONE" --tunnel-through-iap

# Private cluster: forward the cluster API port to localhost:6443 via the bastion.
# After running this, use: oc login https://localhost:6443 -u <user> -p <pass> --insecure-skip-tls-verify=true
# The bastion must be able to resolve and reach the cluster API hostname within the VPC.
.PHONY: example.cluster_private.tunnel
example.cluster_private.tunnel:
	@command -v gcloud >/dev/null 2>&1 || (echo "Error: gcloud CLI not found on PATH."; exit 1)
	@cd examples/cluster_private && \
		BASTION=`terraform output -raw bastion_name` && \
		ZONE=`terraform output -raw bastion_zone` && \
		PROJECT=`terraform output -raw gcp_project_id` && \
		API_URL=`terraform output -raw api_url` && \
		API_HOST=$$(echo "$$API_URL" | sed 's|^https://||' | cut -d: -f1) && \
		echo "Forwarding localhost:6443 -> $$API_HOST:6443 via $$BASTION (Ctrl-C to stop)..." && \
		echo "Then run: oc login https://localhost:6443 -u admin -p <password> --insecure-skip-tls-verify=true" && \
		gcloud compute ssh "$$BASTION" --project="$$PROJECT" --zone="$$ZONE" --tunnel-through-iap \
			-- -L "6443:$$API_HOST:6443" -N

# Dev targets: build provider (binary stays in repo root for dev_overrides), clear lock, re-init.
# Requires dev_overrides in ~/.terraformrc pointing to this repo root (run: make dev-setup).
# Usage: make dev.cluster_with_vpc.apply  make dev.cluster_psc.plan
define dev-targets
.PHONY: dev.$(1) dev.$(1).init dev.$(1).plan dev.$(1).apply dev.$(1).destroy
dev.$(1): dev.$(1).apply

# build (not install) so the binary stays at the repo root where dev_overrides can find it.
# install would move the binary to ~/.terraform.d/plugins, breaking dev_overrides resolution.
dev.$(1).init: build
	@rm -f $(WIF_CONFIG_DIR)/.terraform.lock.hcl examples/$(1)/.terraform.lock.hcl
	@cd $(WIF_CONFIG_DIR) && terraform init -upgrade
	@cd examples/$(1) && terraform init -upgrade

dev.$(1).plan: dev.$(1).init
	@echo "Planning WIF config (dev)..."
	@cd $(WIF_CONFIG_DIR) && terraform plan $(TF_VARS)
	@echo ""
	@echo "Planning examples/$(1) (dev)..."
	@cd examples/$(1) && terraform plan $(TF_VARS) $(EXTRA_TF_VARS)

dev.$(1).apply: dev.$(1).init
	@echo "Step 1: Creating WIF config in OCM (dev)..."
	@cd $(WIF_CONFIG_DIR) && terraform apply -auto-approve $(TF_VARS)
	@echo ""
	@echo "Step 2: Applying examples/$(1) (dev)..."
	@cd examples/$(1) && terraform apply -auto-approve $(TF_VARS) $(EXTRA_TF_VARS)

dev.$(1).destroy: dev.$(1).init
	@echo "Step 1: Destroying examples/$(1) (dev)..."
	@cd examples/$(1) && terraform destroy -auto-approve $(TF_VARS) $(EXTRA_TF_VARS)
	@echo ""
	@echo "Step 2: Destroying WIF config (dev)..."
	@cd $(WIF_CONFIG_DIR) && terraform destroy -auto-approve $(TF_VARS)
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

.PHONY: lint_go
lint_go:
	golangci-lint run ./provider/... ./build/... ./logging/... .

.PHONY: lint_tf
lint_tf: build
	@rm -f $(WIF_CONFIG_DIR)/.terraform.lock.hcl
	@for ex in $(EXAMPLES); do rm -f examples/$$ex/.terraform.lock.hcl; done
	@for mod in $(MODULES); do rm -f modules/$$mod/.terraform.lock.hcl; done
	@echo "Validating terraform/wif_config..."
	@cd $(WIF_CONFIG_DIR) && terraform init -backend=false -input=false -upgrade && terraform validate
	@for ex in $(EXAMPLES); do \
	  echo "Validating examples/$$ex..."; \
	  cd examples/$$ex && terraform init -backend=false -input=false -upgrade && terraform validate && cd $(CURDIR) || exit 1; \
	done
	@for mod in $(MODULES); do \
	  echo "Validating modules/$$mod..."; \
	  cd modules/$$mod && terraform init -backend=false -input=false -upgrade && terraform validate && cd $(CURDIR) || exit 1; \
	done

.PHONY: lint
lint: lint_go lint_tf

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
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2

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
