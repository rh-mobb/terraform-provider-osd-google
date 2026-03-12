# OSD Cluster with Workload Identity Federation (WIF)

This example deploys an OpenShift Dedicated (OSD) cluster using Workload Identity Federation (WIF).
WIF allows OSD to assume GCP service account credentials without storing keys.

## How it works

The Makefile handles the full lifecycle automatically:

1. **`terraform/wif_config/`** — Creates the WIF config in OCM (applied first).
2. **`examples/cluster/`** — Looks up the WIF config by display name, provisions GCP IAM resources, and creates the cluster.

Both configs use the same `cluster_name` variable. The WIF display name is `"${cluster_name}-wif"`, so no IDs need to be passed between runs.

## Assumption: One WIF per cluster

This example follows the recommended practice of **one WIF config per cluster**. While the OCM API allows multiple clusters to share a WIF config, using one per cluster:

- **Version alignment** — WIF configs are version-specific (IAM roles differ per OpenShift version). One per cluster avoids conflicts when clusters upgrade at different times.
- **Lifecycle** — Cluster and WIF config can be created and destroyed together; deleting a cluster does not affect others.
- **Isolation** — Each cluster has its own GCP identities (workload identity pool, service accounts).

## Overview

1. **`data.osdgoogle_wif_config`** — Looks up the existing WIF config by display name (resolved at plan time).
2. **`modules/osd-wif-gcp`** — Provisions GCP resources: workload identity pool, OIDC provider, service accounts, custom roles, IAM bindings.
3. **`osdgoogle_cluster`** — Creates the cluster.

## Prerequisites

- **OCM token** — `OSDGOOGLE_TOKEN` env var or `ocm_token` variable
- **GCP** — Project with WIF prerequisites (see OSD docs); `gcloud auth application-default login`
- **Provider** — `make build` (with `dev_overrides`), `make install`, or use `make dev.cluster.apply` (builds, installs, and runs with local build; no `dev_overrides` needed)

## Apply

From the repository root:

```bash
make example.cluster
```

This applies `terraform/wif_config/` first, then `examples/cluster/` with the same variables. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why WIF config runs in a separate apply. See [terraform/wif_config/README.md](../../terraform/wif_config/README.md) for why the WIF config runs in a separate apply.

For **local provider testing** (build, install to `~/.terraform.d/plugins`, re-init, then run), use the `dev.*` targets. No `dev_overrides` in `~/.terraformrc` needed:

```bash
make dev.cluster.apply
make dev.cluster.plan
make dev.cluster.destroy
```

### Variables

- **`gcp_project_id`** (required) — GCP project ID.
- **`cluster_name`** (default: `my-wif-cluster`) — Must match the value used in `terraform/wif_config/`.
- **`openshift_version`** (default: `4.21.3`) — OpenShift version. Match the cluster version.
- **`admin_password`** (optional) — Cluster admin password. Omit to auto-generate.

## Destroy

From the repository root:

```bash
make example.cluster.destroy
```

Or with the local provider build:

```bash
make dev.cluster.destroy
```

This destroys the cluster first, then the WIF config.
