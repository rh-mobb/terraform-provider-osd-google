# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Removed

- `examples/cluster_psc`: removed — superseded by `examples/cluster_private`, which covers PSC and also sets the cluster API to internal-only.

### Added

- `osdgoogle_cluster`: new `private` boolean attribute — when `true`, sets the OCM API server listening method to `internal`, restricting the cluster API endpoint to private (internal) connectivity only. Requires a BYO VPC (`gcp_network`) and Private Service Connect (`private_service_connect`). Cannot be changed after cluster creation (forces replacement).
- `modules/osd-cluster`: new `private` variable wired to `osdgoogle_cluster.private`.
- `examples/cluster_private`: new example — fully private OSD cluster (API internal-only) with PSC, BYO VPC, and a CentOS Stream 9 bastion VM reachable via `gcloud` IAP SSH tunneling. Includes `make example.cluster_private.ssh` (interactive shell) and `make example.cluster_private.tunnel` (local port-forward for the cluster API) Makefile targets.
- Makefile target `example.<name>.login` — runs `oc login` using `api_url`, `admin_username`, and `admin_password` from the example’s Terraform state (examples that expose those outputs, e.g. `cluster`); does not run `terraform init` (expects the example directory already initialized after apply)
- Makefile targets `wif.init`, `wif.plan`, `wif.apply`, and `wif.destroy` — operate on `terraform/wif_config` only (run from repository root; same `terraform.tfvars` / `TF_VAR_*` conventions as example targets)

### Changed

- Example Makefile targets no longer pass `gcp_project_id` or `cluster_name` on the command line; set them via `TF_VAR_gcp_project_id` / `TF_VAR_cluster_name`, or uncomment values in `terraform.tfvars` (use the same `cluster_name` in `terraform/wif_config` and each example)
- Documented that the provider is experimental, maintained by the Managed OpenShift Black Belt team, and not production-ready (README, provider docs, guides)
- Examples, `terraform/wif_config`, and `osd-cluster` / `osd-wif-config` modules now declare `registry.terraform.io/rh-mobb/osd-google` as the provider source (use `dev_overrides` or dev Makefile targets for local builds)

### Deprecated

### Removed

- `modules/osd-ilb-routing` and `examples/cluster_ilb_routing` — moved to
  [osd-gcp-cudn-routing](https://github.com/rh-mobb/osd-gcp-cudn-routing)

### Fixed

- Makefile `wif.*`, `example.*`, and `dev.*` targets: removed gcloud-based project preflight and fixed `TF_VARS` so Terraform receives optional extra args (`$(TF_VARS)`); variables come only from `terraform.tfvars` / `TF_VAR_*` as documented
- `example.<name>.login` no longer breaks `oc login` with wrong host (`lookup PI` / empty URL): GNU Make was expanding `$(terraform ...)` and treating `$$API` as `$A` + `PI`; the recipe now uses shell backticks only (no Make `$(...)` or `$`-prefixed shell variables in the Makefile)

### Security

## [0.1.0] - 2025-03-13

### Added

- Terraform provider for OpenShift Dedicated (OSD) on GCP
- `osdgoogle_cluster` resource with WIF, PSC, Shared VPC, CMEK support
- `osdgoogle_wif_config` resource for Workload Identity Federation
- `osdgoogle_machine_pool` resource
- `osdgoogle_cluster_admin`, `osdgoogle_cluster_waiter`, `osdgoogle_dns_domain` resources
- Data sources: `osdgoogle_versions`, `osdgoogle_machine_types`, `osdgoogle_regions`, `osdgoogle_wif_config`
- Examples: cluster, cluster_psc, cluster_shared_vpc, cluster_with_vpc, cluster_baremetal, cluster_multi_az (all use WIF)
- Terraform Registry release workflow (GoReleaser, GPG signing)
