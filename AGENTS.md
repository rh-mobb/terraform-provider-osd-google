# AGENTS.md

<!-- keel:start - DO NOT EDIT between these markers -->
## Rules

| Rule | Globs | Always Apply |
|------|-------|--------------|
| agent-behavior | `["**/*"]` | true |
| base | `["**/*"]` | true |
| go | `["**/*.go", "**/go.mod", "**/go.sum"]` | false |
| markdown | `["**/*.md"]` | false |
| python | `["**/*.py", "**/Pipfile", "**/pyproject.toml", "**/requirements*.txt"]` | false |
| scaffolding | `["**/*"]` | true |
| terraform | `["**/*.tf", "**/*.tfvars", "**/*.tfvars.json"]` | false |
| yaml | `["**/*.yaml", "**/*.yml"]` | false |

## Rule Details

### agent-behavior
- **Description:** Universal behavioral safety rules for AI agents interacting with live systems
- **Globs:** `["**/*"]`
- **File:** `.agents/rules/keel/agent-behavior.md`

### base
- **Description:** Global coding standards that apply to all files and languages
- **Globs:** `["**/*"]`
- **File:** `.agents/rules/keel/base.md`

### go
- **Description:** Go coding conventions and best practices
- **Globs:** `["**/*.go", "**/go.mod", "**/go.sum"]`
- **File:** `.agents/rules/keel/go.md`

### markdown
- **Description:** Markdown writing conventions for .md files
- **Globs:** `["**/*.md"]`
- **File:** `.agents/rules/keel/markdown.md`

### python
- **Description:** Python coding conventions and best practices
- **Globs:** `["**/*.py", "**/Pipfile", "**/pyproject.toml", "**/requirements*.txt"]`
- **File:** `.agents/rules/keel/python.md`

### scaffolding
- **Description:** Interactive guidance for essential project scaffolding files
- **Globs:** `["**/*"]`
- **File:** `.agents/rules/keel/scaffolding.md`

### terraform
- **Description:** Best practices and rules for Terraform infrastructure as code
- **Globs:** `["**/*.tf", "**/*.tfvars", "**/*.tfvars.json"]`
- **File:** `.agents/rules/keel/terraform.md`

### yaml
- **Description:** YAML formatting and structure conventions
- **Globs:** `["**/*.yaml", "**/*.yml"]`
- **File:** `.agents/rules/keel/yaml.md`
<!-- keel:end -->

## Local Rules (RHCS-Aligned)

Project-specific rules aligned with Red Hat's terraform-provider-rhcs. Local rules take precedence over keel rules for overlapping topics.

| Rule | Globs | Always Apply |
|------|-------|--------------|
| terraform-provider | `["**/*.go"]` | false |
| testing | `["**/*_test.go", "**/subsystem/**"]` | false |
| project | `["**/*"]` | true |

### terraform-provider
- **Description:** Terraform provider patterns aligned with Red Hat RHCS conventions
- **Globs:** `["**/*.go"]`
- **File:** `.agents/rules/local/terraform-provider.md`

### testing
- **Description:** Ginkgo/Gomega and subsystem test patterns aligned with Red Hat RHCS conventions
- **Globs:** `["**/*_test.go", "**/subsystem/**"]`
- **File:** `.agents/rules/local/testing.md`

### project
- **Description:** Project conventions, build workflow, and commit format aligned with Red Hat RHCS
- **Globs:** `["**/*"]`
- **File:** `.agents/rules/local/project.md`

## References

The `references/` folder contains upstream source repositories and API specifications cloned locally to provide AI agents with rich, offline context. This folder is gitignored and must be cloned manually — see the [README](README.md#ai-agent-development) for instructions.

| Reference | What it is | Useful for |
|-----------|------------|------------|
| `references/OCM.json` | OpenAPI 3.0 spec for the OCM Cluster Management API (`clusters_mgmt/v1`, 153 endpoints) | Understanding the exact shape of OCM API requests and responses; discovering available fields, enums, and nested types without needing a running server |
| `references/ocm-sdk-go/` | Go SDK used to call the OCM API (`github.com/openshift-online/ocm-sdk-go`) | The provider imports this SDK directly — agents can navigate builder types, client methods, and type aliases to understand how to construct or parse OCM API calls |
| `references/ocm-cli/` | Source for the `ocm` CLI tool that wraps the OCM API | Cross-referencing how the CLI models clusters, WIF configs, and other resources; useful for understanding field semantics and validation rules |
| `references/terraform-provider-rhcs/` | Red Hat's official Terraform provider for ROSA (OpenShift on AWS) | The structural template for this provider — agent can study resource schemas, state management patterns, subsystem test conventions, and helper utilities to replicate the same patterns for GCP |
| `references/terraform-google-osd/` | Terraform modules for deploying OSD on GCP (VPC, PSC, WIF, service accounts) | Understanding the GCP-side infrastructure that OSD clusters depend on; useful when implementing or debugging network-related provider fields |
