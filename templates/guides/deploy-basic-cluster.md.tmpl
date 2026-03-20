---
page_title: "Deploy a basic OSD cluster"
subcategory: ""
description: |-
  Deploy an OpenShift Dedicated (OSD) cluster on GCP using default networking and the osd-wif-config and osd-cluster modules.
---

# Deploy a Basic OSD Cluster

This guide walks through deploying a basic OSD cluster on Google Cloud Platform with default networking. The cluster uses Workload Identity Federation (WIF) for secure, keyless authentication between OCM and your GCP project.

> **Note:** The osdgoogle provider is **experimental** software from the **Red Hat Managed OpenShift Black Belt** team. It is **not** production-ready and is **not** a supported Red Hat product.

## Prerequisites

Before you begin, ensure you have:

* An [OCM offline token](https://console.redhat.com/openshift/token/rosa)
* A GCP project with billing enabled and the [required APIs](https://cloud.google.com/iam/docs/workload-identity-federation) for WIF
* Application Default Credentials configured (`gcloud auth application-default login`)
* [Minimum GCP IAM permissions](../#minimum-gcp-iam-permissions): `roles/iam.workloadIdentityPoolAdmin`, `roles/iam.serviceAccountAdmin`, `roles/iam.roleAdmin`, `roles/resourcemanager.projectIamAdmin`, and `roles/resourcemanager.projectViewer`

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
  openshift_version = var.openshift_version
}

provider "google" {
  project = var.gcp_project_id
}
```

Create `variables.tf`:

```terraform
variable "ocm_token" {
  type        = string
  sensitive   = true
  default     = ""
  description = "OCM offline token (use OSDGOOGLE_TOKEN env var when empty)"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

variable "cluster_name" {
  type        = string
  default     = "my-cluster"
  description = "Name of the cluster"
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z)"
}

variable "admin_password" {
  type        = string
  sensitive   = true
  default     = ""
  description = "Cluster admin password (optional; auto-generated if omitted)"
}
```

## Step 2: Create WIF config

Create the WIF configuration in OCM. The cluster name used here must match the cluster module so the cluster can look up the WIF config by display name (`{cluster_name}-wif`).

```terraform
module "wif_config" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-wif-config"

  gcp_project_id    = var.gcp_project_id
  cluster_name      = var.cluster_name
  openshift_version = var.openshift_version
}
```

## Step 3: Create the cluster

The `osd-cluster` module provisions the GCP IAM resources (via the internal `osd-wif-gcp` module) and creates the OSD cluster. It looks up the WIF config by display name, which defaults to `{cluster_name}-wif`.

```terraform
module "cluster" {
  source = "github.com/rh-mobb/terraform-provider-osd-google//modules/osd-cluster"

  name              = var.cluster_name
  cloud_region      = "us-central1"
  gcp_project_id    = var.gcp_project_id
  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  create_admin   = true
  admin_password = var.admin_password != "" ? var.admin_password : null
}
```

## Step 4: Apply

Initialize and apply:

```console
$ terraform init
$ export OSDGOOGLE_TOKEN="your-offline-token"
$ terraform plan -var="gcp_project_id=my-gcp-project"
$ terraform apply -var="gcp_project_id=my-gcp-project"
```

Or use a `terraform.tfvars` file (ensure it is gitignored if it contains secrets):

```hcl
gcp_project_id = "my-gcp-project"
cluster_name   = "my-cluster"
openshift_version = "4.21.3"
```

## Outputs

Add `outputs.tf` to expose cluster details:

```terraform
output "cluster_id" {
  value     = module.cluster.cluster_id
  sensitive = false
}

output "api_url" {
  value     = module.cluster.api_url
  sensitive = false
}

output "console_url" {
  value     = module.cluster.console_url
  sensitive = false
}

output "admin_username" {
  value     = module.cluster.admin_username
  sensitive = false
}

output "admin_password" {
  value     = module.cluster.admin_password
  sensitive = true
}
```

## Next steps

* [Deploy a BYO VPC cluster](deploy-byo-vpc-cluster.md) for custom networking
* Add [machine pools](resources/machine_pool.md) for additional compute capacity
* Configure [cluster admin](resources/cluster_admin.md) or other identity providers
