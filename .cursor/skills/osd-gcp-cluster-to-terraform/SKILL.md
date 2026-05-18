---
name: osd-gcp-cluster-to-terraform
description: Reverse-engineers an existing OpenShift Dedicated on GCP deployment into production-style Terraform using this repo’s examples and modules, scopes what should be reproducible vs imported-only via a short questionnaire, emits variable-driven roots (tfvars per cluster for N clusters), and documents terraform import and operational caveats. Use when adopting an existing OSD-on-GCP cluster, generating Terraform from live OCM/GCP state, multi-cluster tfvars layouts, or when the user asks for cluster-to-code or import playbooks for rh-mobb/osd-google.
---

# OSD on GCP — cluster to Terraform (adoption codegen)

## When to use this skill

Apply when the user wants to **represent an already-running OSD-on-GCP stack in Terraform** (or a greenfield twin), using **this repository’s provider, examples, and modules**, with:

- **Discovery** from OCM + GCP (and optionally cluster metadata),
- **Explicit scoping** (what to manage vs import-only),
- **Outputs** that are **parameterized** (`variables` + per-environment `*.tfvars`) so **multiple clusters** do not hardcode conflicting names,
- **Import instructions** alongside any generated root module.

If the workspace is not `terraform-provider-osd-google` (or a fork with the same layout), say so and either adapt paths or stop.

## Core constraints (non-negotiable)

1. **Prefer this repo as the source of patterns** — start from `examples/` (e.g. `examples/cluster`, `cluster_with_vpc`, `cluster_shared_vpc`, `cluster_psc`, `cluster_multi_az`, `cluster_baremetal`) and `modules/` (`osd-cluster`, `osd-vpc`, `osd-wif-gcp`, `osd-wif-config`). Read the closest match before inventing structure.
2. **No literal production names in root `.tf`** — project IDs, cluster names, regions, VPC names, subnet names, and bucket/pool prefixes must flow through **variables**; each cluster gets a **`terraform.tfvars` or `env/foo.tfvars`** file (or `tfvars` per workspace). Locals may derive prefixed resource names from a single `deployment_name` / `cluster_name` / `name_prefix` variable.
3. **Split OCM vs GCP responsibility** — `osdgoogle_*` resources talk to **OCM**; VPC, subnets, PSC attachments, IAM, and most GCP objects use the **`google` provider** (often via modules here). Generated docs must say which state moves with which provider.
4. **Secrets never committed** — service account keys, tokens, kubeconfigs: variables marked `sensitive`, `tfvars` gitignored, instructions to use env vars or a secret backend.
5. **Honest limits** — the provider’s `osdgoogle_cluster` **Read** path does not repopulate every optional block from OCM; **import** IDs differ per resource. Do not promise bit-identical state round-trip without verification. Point agents to [reference.md](reference.md) for import formats and caveats.

## Workflow (agent)

### 1. Establish inputs

Collect (or instruct how to obtain):

| Source | What |
|--------|------|
| OCM | Cluster ID, name, region, multi-AZ, version, CCS/WIF IDs, machine pools (non-default), DNS domain objects if relevant |
| GCP | Project(s), VPC name, subnets (control plane / compute / PSC if used), Shared VPC host project if applicable, PSC config, KMS if CMEK |

If the user cannot run `gcloud`/`ocm`, ask for exports (sanitized) or read-only descriptions.

### 2. Questionnaire — “how replicable?”

Ask briefly (adapt wording); defaults if unanswered: **manage networking + WIF in TF, import cluster**.

Suggested prompts:

- **Greenfield twin vs adopt-only?** (Same topology new names vs bind existing IDs.)
- **Manage VPC/subnets/PSC in Terraform?** Or assume existing `google` module state?
- **Manage WIF (`osdgoogle_wif_config` / `modules/osd-wif-*`) in TF?** Or import-only?
- **Manage `osdgoogle_cluster` lifecycle in TF?** (Recommend `lifecycle { prevent_destroy = true }` until plans are trusted.)
- **Machine pools:** only non-default pools via `osdgoogle_machine_pool` (default worker pool has provider limitations — see reference).
- **Cluster admin / IdP:** include `osdgoogle_cluster_admin` only if they want TF to own it.
- **CI/backend:** remote state bucket, workspace-per-cluster vs single root + multiple tfvars?

### 3. Select template

Map topology → example/module set:

| Topology | Primary references |
|----------|-------------------|
| Simple CCS | `examples/cluster`, `modules/osd-cluster` |
| BYO VPC | `examples/cluster_with_vpc`, `modules/osd-vpc` |
| Shared VPC | `examples/cluster_shared_vpc` |
| PSC | `examples/cluster_psc`, `modules/osd-vpc` PSC pieces |
| Multi-AZ | `examples/cluster_multi_az` |
| Bare metal | `examples/cluster_baremetal` |

Read `README.md` under the chosen example and matching `modules/*/README.md` before generating files.

### 4. Generate artifacts

Produce a **coherent root module** (or `clusters/<name>/` subdirs if they asked for many roots) with:

- `variables.tf` — typed variables with validation where helpful; a **small** set of “identity” vars (`project_id`, `region`, `cluster_display_name`, `name_prefix`, etc.).
- `terraform.tfvars.example` — commented placeholders; **no secrets**.
- `providers.tf` / `versions.tf` — pin `osdgoogle` and `google` (and `random` if used) per repo conventions.
- `main.tf` — wire modules with `var.*` only; use `locals` for derived names (e.g. `"${var.name_prefix}-subnet"`).
- `outputs.tf` — OCM cluster id, endpoints, submodule outputs needed for imports or handoff.
- **`IMPORT.md`** (or section in `README.md`) — ordered import commands / `import` blocks; note **OCM cluster ID** vs GCP resource IDs; link [reference.md](reference.md).

For **N clusters**: one root module pattern + **`cluster-a.tfvars`**, **`cluster-b.tfvars`**, … (or workspaces) where only variable values change; resource naming must include a **per-cluster prefix** from variables.

### 5. Verification checklist (tell the user)

- `terraform fmt` / `terraform validate`
- `terraform plan` with **import already done** or **fresh** account — expectations differ; spell out which.
- Confirm **destroy** would not touch production until `prevent_destroy` is removed.

## Tone and scope

- Prefer **actionable file trees and copy-paste blocks** over long prose.
- If discovery data is incomplete, **generate conservative TF** (fewer managed resources + more “data source + import” notes) rather than guessing CIDRs or WIF role names.
- **Do not** run destructive GCP/OCM operations as part of skill execution unless the user explicitly orders them.

## Additional resources

- Minimal **cluster-only** `terraform import` (no WIF/VPC): provider guide [docs/guides/adopt-existing-cluster.md](../../../docs/guides/adopt-existing-cluster.md) (published from `templates/guides/adopt-existing-cluster.md.tmpl` after `make docs`).
- Import IDs, resource scope, and known provider gaps: [reference.md](reference.md)
