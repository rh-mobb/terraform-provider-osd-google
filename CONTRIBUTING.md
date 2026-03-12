# Contributing to Terraform Provider for OSD on Google Cloud

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Reporting Bugs

- Use [GitHub Issues](https://github.com/rh-mobb/terraform-provider-osd-google/issues) to report bugs.
- Include the Terraform and provider versions, relevant configuration (redact secrets), and error output.
- Describe expected vs actual behavior.

## Proposing Features

- Open a [GitHub Issue](https://github.com/rh-mobb/terraform-provider-osd-google/issues) with the `enhancement` label.
- Describe the use case and proposed API (resource/attribute changes).
- Reference OCM API or Red Hat documentation when relevant.

## Submitting Pull Requests

1. **Fork** the repository and create a branch from `main`.
2. **Develop** your changes following the project's conventions (see [AGENTS.md](AGENTS.md) for AI agent guidance).
3. **Test** locally:
   ```bash
   make unit-test
   make fmt
   make build
   make docs   # If you changed schema
   ```
4. **Commit** with clear, imperative messages (e.g., "Add private cluster support").
5. **Push** your branch and open a PR against `main`.
6. Ensure CI passes and address review feedback.

## Development Setup

- **Requirements:** Go 1.24+, Terraform 1.0+, [jq](https://jqlang.github.io/jq/)
- **Build:** `make build`
- **Install locally:** `make install`
- **Run examples with local build:** `make dev.cluster.apply`, `make dev.cluster.plan`, `make dev.cluster.destroy` — these install the provider, clear locks, re-init, and run. No `dev_overrides` in `~/.terraformrc` required.
- **Run tests:** `make unit-test`, `make subsystem-test` (requires OCM mock)
- **Format:** `make fmt`

## Code Style

- **Go:** Use `gofmt`; run `make fmt_go`.
- **Terraform:** Use `terraform fmt`; run `make fmt_tf`.
- Follow conventions in existing code and [AGENTS.md](AGENTS.md).

## Code of Conduct

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).
