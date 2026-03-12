# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Terraform provider for OpenShift Dedicated (OSD) on GCP
- `osdgoogle_cluster` resource with WIF, PSC, Shared VPC, CMEK support
- `osdgoogle_wif_config` resource for Workload Identity Federation
- `osdgoogle_machine_pool` resource
- Data sources: `osdgoogle_versions`, `osdgoogle_machine_types`, `osdgoogle_regions`
- Examples: cluster, cluster_psc, cluster_shared_vpc, cluster_with_vpc (all use WIF)

### Changed

- (none yet)

### Deprecated

- (none yet)

### Removed

- (none yet)

### Fixed

- (none yet)

### Security

- (none yet)
