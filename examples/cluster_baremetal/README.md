# OSD Cluster with Bare Metal as Default Compute (Single AZ)

This example deploys an OpenShift Dedicated (OSD) cluster using bare metal instances (e.g. `c3-standard-192-metal`) as the default worker node type. The cluster is single-AZ. OCM supports multiple zones for bare metal; however, the machine type must be available in each zone you specify. GCP bare metal types are zone-specific—e.g. `c3-standard-192-metal` is available in `us-central1-a` but not in `us-central1-b`. Use `google_compute_machine_types` or GCP docs to verify zone availability before specifying `availability_zones`.

**NOTE:** Secure Boot (Shielded VMs) is not supported on bare metal instance types.

## How it works

1. **`terraform/wif_config/`** — Creates the WIF config in OCM (applied first by Makefile).
2. **`examples/cluster_baremetal/`** — Looks up the WIF config, provisions GCP IAM, and creates the cluster with `compute_machine_type` and `availability_zones` set for bare metal.

Both configs use the same `cluster_name` variable. The WIF display name is `"${cluster_name}-wif"`.

## Overview

1. **`data.osdgoogle_wif_config`** — Looks up the WIF config by display name.
2. **`data.osdgoogle_machine_types`** — Validates instance type availability (OCM global catalog, no GCP creds).
3. **`modules/osd-wif-gcp`** — Provisions GCP IAM for WIF.
4. **`osdgoogle_cluster`** — Creates the cluster with bare metal as default compute, single zone.

## Prerequisites

- **OCM token** — `OSDGOOGLE_TOKEN` env var or `ocm_token` variable
- **GCP** — Project with WIF prerequisites; `gcloud auth application-default login`
- **Provider** — `make build` (with `dev_overrides`) or `make install`

## Apply

```bash
make example.cluster_baremetal
```

### Variables

- **`gcp_project_id`** (required) — GCP project ID.
- **`cluster_name`** (default: `my-baremetal-cluster`) — Must match `terraform/wif_config`.
- **`openshift_version`** (default: `4.21.3`) — OpenShift version.
- **`admin_password`** (optional) — Cluster admin password. Omit to auto-generate.
- **`gcp_region`** (default: `us-central1`) — GCP region.
- **`compute_machine_type`** (default: `c3-standard-192-metal`) — Bare metal machine type.
- **`compute_nodes`** (default: `3`) — Number of worker nodes.
- **`availability_zone`** (default: `us-central1-a`) — Single zone. The machine type must be available in this zone (e.g. c3-standard-192-metal is in us-central1-a but not us-central1-b).

## Destroy

```bash
make example.cluster_baremetal.destroy
```
