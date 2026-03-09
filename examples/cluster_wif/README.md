# OSD Cluster with Workload Identity Federation (WIF)

This example deploys an OpenShift Dedicated (OSD) cluster using Workload Identity Federation (WIF).
WIF allows OSD to assume GCP service account credentials without storing keys.

## Assumption: One WIF per cluster

This example follows the recommended practice of **one WIF config per cluster**. While the OCM API allows multiple clusters to share a WIF config, using one per cluster:

- **Version alignment** — WIF configs are version-specific (IAM roles differ per OpenShift version). One per cluster avoids conflicts when clusters upgrade at different times.
- **Lifecycle** — Cluster and WIF config can be created and destroyed together; deleting a cluster does not affect others.
- **Isolation** — Each cluster has its own GCP identities (workload identity pool, service accounts).

The example uses `display_name = "${cluster_name}-wif"` to keep the WIF config clearly tied to its cluster.

## Overview

1. **`osdgoogle_wif_config`** — Registers the WIF config in OCM; OCM returns a blueprint (pool, service accounts, IAM).
2. **`modules/osd-wif-gcp`** — Provisions GCP resources: workload identity pool, OIDC provider, service accounts, custom roles, IAM bindings.
3. **`osdgoogle_cluster`** — Creates the cluster, depending on the GCP resources.

## Why Two-Phase Apply?

The GCP module uses `for_each` over `service_accounts` and `support` from the WIF config blueprint.
Terraform requires `for_each` keys to be known at **plan time**, but the blueprint is only available **after** the WIF config is created in OCM.

Therefore, apply must run in two phases:

| Phase | Target | What happens |
|-------|--------|--------------|
| **1** | `osdgoogle_wif_config.wif` | Creates WIF config in OCM; OCM returns the blueprint. State now includes `service_accounts`, `pool_id`, etc. |
| **2** | *(full)* | Plan can now resolve `for_each`; creates GCP pool, SAs, IAM, and cluster. |

## Prerequisites

- **OCM token** — `OSDGOOGLE_TOKEN` env var or `ocm_token` variable
- **GCP** — Project with WIF prerequisites (see OSD docs); `gcloud auth application-default login`
- **Provider** — `make build` (with `dev_overrides`) or `make install`

## Apply

### Option A: Makefile target (recommended)

From the repository root:

```bash
make apply-wif-cluster
```

This runs both phases in sequence with `-auto-approve` (non-interactive).

### Option B: Manual two-phase apply

```bash
cd examples/cluster_wif
terraform init   # if not using dev_overrides

# Phase 1: Create WIF config; OCM returns blueprint
terraform apply -target=osdgoogle_wif_config.wif

# Phase 2: Create GCP resources and cluster
terraform apply
```

### Variables

Set `gcp_project_id` (required) and optionally `cluster_name`, `role_prefix`, `openshift_version`:

- **`openshift_version`** (default: `"4.21.3"`) — OpenShift version (x.y.z). Scopes WIF IAM resources; patch (.z) is stripped for role templates. Match the cluster version.

```bash
terraform apply -var="gcp_project_id=my-gcp-project"
# or use terraform.tfvars
```

## Other targets

- **`make plan-wif-cluster`** — Run `terraform plan` in the WIF example (shows Phase 2 plan after Phase 1 is applied).
- **`make destroy-wif-cluster`** — Destroy the WIF example. Uses `-refresh=false` so Terraform evaluates `for_each` from existing state instead of refreshed values (avoids "known only after apply" during destroy).
