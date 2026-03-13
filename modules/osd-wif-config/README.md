# OSD WIF Config Module

Creates the Workload Identity Federation (WIF) config in OCM for OpenShift Dedicated (OSD) clusters on GCP. This module encapsulates the OCM-side WIF resource and is used by `terraform/wif_config/` as Phase 1 of the two-phase apply workflow.

## Resources

- **data.google_project** - Fetches GCP project number
- **osdgoogle_wif_config** - Creates WIF config in OCM

## Usage

```hcl
module "wif_config" {
  source = "../../modules/osd-wif-config"

  gcp_project_id    = var.gcp_project_id
  cluster_name      = var.cluster_name
  openshift_version = var.openshift_version  # optional, default 4.21.3
  role_prefix       = var.role_prefix        # optional
}
```

## Variables

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| gcp_project_id | string | yes | - | GCP project ID |
| cluster_name | string | yes | - | Base name; WIF display name is `{cluster_name}-wif` |
| openshift_version | string | no | "4.21.3" | OpenShift version (x.y.z); WIF roles use x.y only |
| role_prefix | string | no | null | Prefix for custom IAM roles; defaults to cluster_name with hyphens/underscores stripped |

## Outputs

| Name | Description |
|------|-------------|
| wif_config_id | OCM WIF config ID |
| display_name | WIF config display name (used by cluster module data source lookup) |

## Two-phase workflow

The WIF config must exist in OCM **before** the cluster module can run. The cluster module uses `data.osdgoogle_wif_config` to look up the config by display name (`{cluster_name}-wif`) and provision GCP IAM.

From the repo root:

```bash
make example.cluster
```

This applies `terraform/wif_config/` first (which uses this module), then the example.
