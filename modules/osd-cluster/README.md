# OSD Cluster Module

Creates an OpenShift Dedicated (OSD) cluster on GCP with Workload Identity Federation (WIF). Handles the full cluster lifecycle: WIF GCP provisioning, cluster creation, optional admin user, and optional machine pools.

Internally nests the [osd-wif-gcp](../osd-wif-gcp) module and provisions GCP IAM before creating the cluster.

## Prerequisites

- WIF config must exist in OCM (e.g. via `terraform/wif_config/` which uses [osd-wif-config](../osd-wif-config))
- The cluster module looks up the WIF config by display name (`{name}-wif` by default)

## Usage

```hcl
module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  create_admin   = true
  admin_password = var.admin_password  # optional; auto-generated if null

  machine_pools = var.machine_pools
}
```

With BYO VPC and PSC:

```hcl
module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  private_service_connect = var.enable_psc ? {
    service_attachment_subnet = module.osd_vpc.psc_subnet
  } : null

  security = var.enable_psc ? { secure_boot = true } : null

  machine_pools = var.machine_pools
}
```

Multi-AZ:

```hcl
module "cluster" {
  source = "../../modules/osd-cluster"

  name               = var.cluster_name
  cloud_region       = var.gcp_region
  gcp_project_id     = var.gcp_project_id
  multi_az           = true
  availability_zones = slice(data.google_compute_zones.available.names, 0, 3)

  gcp_network = { ... }
  machine_pools = var.machine_pools
}
```

## Key variables

| Name | Type | Required | Description |
|------|------|----------|-------------|
| name | string | yes | Cluster name; WIF lookup uses `{name}-wif` |
| cloud_region | string | yes | GCP region |
| gcp_project_id | string | yes | GCP project ID |
| wif_config_display_name | string | no | Override WIF lookup; defaults to `{name}-wif` |
| openshift_version | string | no | OpenShift version (default 4.21.3) |
| compute_nodes | number | no | Default worker count (default 3) |
| compute_machine_type | string | no | Machine type for default pool |
| multi_az | bool | no | Deploy across multiple AZs |
| availability_zones | list(string) | no | GCP zones (1 for single-AZ, 3 for multi-AZ) |
| gcp_network | object | no | BYO VPC (vpc_name, compute_subnet, control_plane_subnet, vpc_project_id) |
| private_service_connect | object | no | PSC config (service_attachment_subnet) |
| security | object | no | Security options (secure_boot) |
| create_admin | bool | no | Create HTPasswd IDP and cluster-admin (default false) |
| admin_username | string | no | Admin username (default admin) |
| admin_password | string | no | Admin password; auto-generated if null when create_admin=true |
| machine_pools | list(object) | no | Additional machine pools |

See [variables.tf](variables.tf) for the full variable schema.

## Outputs

| Name | Description |
|------|-------------|
| cluster_id | OCM cluster ID |
| wif_config_id | WIF config ID |
| api_url | Kubernetes API URL |
| console_url | OpenShift web console URL |
| domain | Cluster base domain |
| state | Cluster state |
| infra_id | Infrastructure ID |
| admin_username | Admin username (when create_admin=true) |
| admin_password | Admin password (sensitive; when create_admin=true) |
