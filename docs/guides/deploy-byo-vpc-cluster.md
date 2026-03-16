---
page_title: "Deploy a BYO VPC OSD cluster"
subcategory: ""
description: |-
  Deploy an OSD cluster on GCP with a bring-your-own VPC, using the osd-vpc, osd-wif-config, and osd-cluster modules.
---

# Deploy a BYO VPC cluster

This guide shows how to deploy a Red Hat OpenShift Dedicated (OSD) cluster on GCP with a module-managed VPC. The `osd-vpc` module creates the VPC, subnets, Cloud NAT, and optional Private Service Connect (PSC) resources. The `osd-wif-config` and `osd-cluster` modules handle Workload Identity Federation and cluster creation.

## Prerequisites

* **OCM token** — Generate an offline token at [console.redhat.com/openshift/token/rosa](https://console.redhat.com/openshift/token/rosa)
* **GCP project** — With billing enabled and APIs required for OSD
* **GCP credentials** — `gcloud auth application-default login` for the identity that will run Terraform
* **Minimum GCP permissions** — `roles/iam.workloadIdentityPoolAdmin`, `roles/iam.serviceAccountAdmin`, `roles/iam.roleAdmin`, `roles/resourcemanager.projectIamAdmin`, `roles/compute.networkAdmin`, `roles/compute.routerAdmin`. Add `roles/dns.admin` if using PSC.

## Step 1: Provider configuration

Create `providers.tf`:

```terraform
terraform {
  required_providers {
    osdgoogle = {
      source  = "rh-mobb/osd-google"
      version = ">= 0.1.0"
    }
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
  }
}

provider "osdgoogle" {
  token             = var.ocm_token != "" ? var.ocm_token : null
  openshift_version  = var.openshift_version
}

provider "google" {
  project = var.gcp_project_id
}
```

Use `OSDGOOGLE_TOKEN` environment variable instead of `var.ocm_token` when possible to avoid storing secrets in variable files.

## Step 2: Create WIF config

The `osd-wif-config` module creates the Workload Identity Federation configuration in OCM. The cluster module uses this to provision GCP IAM resources.

```terraform
module "wif_config" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-wif-config"

  gcp_project_id    = var.gcp_project_id
  cluster_name      = var.cluster_name
  openshift_version = var.openshift_version
  role_prefix       = var.role_prefix
}
```

## Step 3: Create VPC

The `osd-vpc` module creates a VPC with control plane and compute subnets, Cloud Router, Cloud NAT, and optional PSC resources for private clusters.

```terraform
module "osd_vpc" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-vpc"

  project_id             = var.gcp_project_id
  region                 = var.gcp_region
  cluster_name           = var.cluster_name
  enable_psc             = var.enable_psc
  enable_private_cluster = var.enable_psc
}
```

## Step 4: Create cluster

The `osd-cluster` module provisions GCP IAM (via the WIF blueprint), then creates the OSD cluster. Pass the VPC outputs for BYO VPC networking.

```terraform
module "cluster" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  private_service_connect = var.enable_psc ? {
    service_attachment_subnet = module.osd_vpc.psc_subnet
  } : null

  security = var.enable_psc ? {
    secure_boot = true
  } : null

  create_admin   = true
  admin_password = var.admin_password != "" ? var.admin_password : null

  machine_pools = var.machine_pools
}
```

## Optional: Enable Private Service Connect (PSC)

Private Service Connect provides private connectivity to Google APIs and is required for fully private clusters. Set `enable_psc = true` and ensure your OpenShift version supports PSC (4.17+). The `osd-vpc` module creates the PSC subnet, global address, and private DNS zones. The cluster module must receive `service_attachment_subnet` and should enable `secure_boot` when using PSC.

## Complete example

`main.tf` combining all modules:

```terraform
module "wif_config" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-wif-config"

  gcp_project_id    = var.gcp_project_id
  cluster_name      = var.cluster_name
  openshift_version = var.openshift_version
}

module "osd_vpc" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-vpc"

  project_id             = var.gcp_project_id
  region                 = var.gcp_region
  cluster_name           = var.cluster_name
  enable_psc             = var.enable_psc
  enable_private_cluster = var.enable_psc
}

module "cluster" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  private_service_connect = var.enable_psc ? {
    service_attachment_subnet = module.osd_vpc.psc_subnet
  } : null

  security = var.enable_psc ? {
    secure_boot = true
  } : null

  create_admin   = true
  admin_password = var.admin_password != "" ? var.admin_password : null

  machine_pools = var.machine_pools
}
```

## Variables

Create `variables.tf` with at least:

```terraform
variable "gcp_project_id" {
  type        = string
  description = "GCP project ID"
}

variable "cluster_name" {
  type        = string
  default     = "my-vpc-cluster"
  description = "Cluster name"
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z)"
}

variable "ocm_token" {
  type        = string
  sensitive   = true
  default     = ""
  description = "OCM offline token (or use OSDGOOGLE_TOKEN env var)"
}

variable "gcp_region" {
  type        = string
  default     = "us-central1"
  description = "GCP region"
}

variable "enable_psc" {
  type        = bool
  default     = false
  description = "Enable Private Service Connect for private cluster (OpenShift 4.17+)"
}

variable "admin_password" {
  type        = string
  sensitive   = true
  default     = ""
  description = "Cluster admin password (optional; auto-generated if omitted)"
}

variable "role_prefix" {
  type        = string
  default     = null
  nullable    = true
  description = "Prefix for GCP IAM roles. Defaults to cluster_name."
}

variable "machine_pools" {
  type        = list(any)
  default     = []
  description = "Additional machine pools"
}
```

## Apply

```console
$ terraform init
$ terraform plan -var-file=terraform.tfvars
$ terraform apply -var-file=terraform.tfvars
```

## Outputs

Add `outputs.tf` to expose cluster endpoints:

```terraform
output "api_url" {
  value = module.cluster.api_url
}

output "console_url" {
  value = module.cluster.console_url
}

output "admin_username" {
  value     = module.cluster.admin_username
  sensitive = true
}

output "admin_password" {
  value     = module.cluster.admin_password
  sensitive = true
}
```

## Next steps

* Add machine pools via the `machine_pools` variable or the `osdgoogle_machine_pool` resource
* Use a shared VPC by providing `vpc_project_id` in `gcp_network` and creating subnets in the host project
* See the [cluster_with_vpc example](https://github.com/rh-mobb/terraform-provider-osd-google/tree/main/examples/cluster_with_vpc) for a full reference
