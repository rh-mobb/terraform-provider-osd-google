---
description: "Project conventions, build workflow, and commit format aligned with Red Hat RHCS"
globs: ["**/*"]
alwaysApply: true
---

# Project Conventions (RHCS-Aligned)

Align with Red Hat's terraform-provider-rhcs for seamless handoff.

## Commit Format

RHCS uses structured commit messages:

```
[JIRA-TICKET] | [TYPE]: <MESSAGE>

[optional BODY]
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`

- `fix:` — patches a bug
- `feat:` — introduces a new feature
- Others: `build`, `ci`, `docs`, `perf`, `refactor`, `style`, `test`

## Operational Safety

**Never run the following commands without explicit user permission:**

- `make` (any target) — builds, installs, or runs tests that modify the system
- `ocm delete`, `ocm create` — mutates live OCM resources (clusters, WIF configs)
- `terraform apply`, `terraform destroy` — mutates cloud infrastructure
- `gcloud` write operations — mutates GCP resources

Always **ask the user to run** these commands and provide the exact command to copy-paste. Read-only commands (`ocm get`, `ocm list`, `ocm describe`, `terraform state list`, `gcloud ... list`) are safe to run directly.

## Build and Makefile

- `make build` — compile with version/commit ldflags
- `make dev-setup` — build and print `dev_overrides` config for `~/.terraformrc`
- `make install` — build and install to `~/.terraform.d/plugins/terraform.local/local/<provider>/`
- `make unit-test` — Ginkgo over `provider/` and `internal/`
- `make subsystem-test` — depends on `make install`; runs Ginkgo in `subsystem/`
- `make test` / `make tests` — unit + subsystem
- `make fmt` — format Go and Terraform files
- `make generate` — run `go generate ./...` (docs, mocks)
- `make check-gen` — ensure generated files are committed (CI)

## Development Workflow

Preferred local dev flow uses Terraform `dev_overrides` (no `terraform init` needed):

1. `make dev-setup` — builds binary and prints `~/.terraformrc` config
2. Add the `dev_overrides` block to `~/.terraformrc` (one-time)
3. `make build` — rebuild after code changes
4. `terraform plan` / `terraform apply` in any `examples/` directory

Examples use `registry.terraform.io/rh-mobb/osd-google` as the provider source.
The `dev_overrides` block redirects Terraform to the local binary.

**Build settings**:
- `CGO_ENABLED=0` for static binaries
- Version and commit injected via `-ldflags`

## Documentation

- Use `tfplugindocs` via `go:generate` to generate provider docs
- Keep examples in `examples/` per resource
- Run `make docs` (or equivalent) after schema changes

## Code Style

- Use `make fmt_go` and `make fmt_tf` before committing
- Follow Effective Go and Terraform Plugin Framework docs
