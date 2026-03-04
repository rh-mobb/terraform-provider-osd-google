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
