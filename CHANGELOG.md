# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

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
