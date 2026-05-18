---
page_title: "Adopt an existing OSD cluster (import only)"
subcategory: ""
description: |-
  Import an existing OpenShift Dedicated on GCP cluster into Terraform using osdgoogle_cluster only; does not cover WIF, VPC, or other resources.
---

# Adopt an Existing OSD on GCP Cluster (Import Only)

This guide covers the **minimal** path: bring an **already running** OpenShift Dedicated (OSD) on Google Cloud cluster under Terraform by **importing** the `osdgoogle_cluster` resource. That is the traditional `terraform import` workflow against the OpenShift Cluster Manager (OCM) API.

> **Note:** The osdgoogle provider is **experimental** software from the **Red Hat Managed OpenShift Black Belt** team. It is **not** production-ready and is **not** a supported Red Hat product.

## What this guide does **not** cover

Importing **`osdgoogle_cluster`** only registers the **cluster object in OCM** in Terraform state. It does **not**:

* Import or manage **`osdgoogle_wif_config`** (Workload Identity Federation configuration in OCM)
* Model **VPCs, subnets, PSC, firewall rules, Cloud NAT**, or other **Google provider** resources
* Replace your need for **GCP IAM**, **DNS**, or **support tickets** outside Terraform

Those pieces may already exist in your GCP project and in OCM; this guide does not walk through importing or generating Terraform for them.

### Broader adoption (WIF, VPC, modules, reproducible roots)

For a **more exhaustive** adoption—reverse-engineering networking, WIF, variables, multiple clusters, and import instructions together—use the **Cursor project skill** shipped with this repository (or follow the same ideas manually):

* Skill: [`.cursor/skills/osd-gcp-cluster-to-terraform/SKILL.md`](https://github.com/rh-mobb/terraform-provider-osd-google/blob/main/.cursor/skills/osd-gcp-cluster-to-terraform/SKILL.md) (clone the repo; path is under the repository root)
* Human-oriented pointer: [README — AI Agent Development](https://github.com/rh-mobb/terraform-provider-osd-google/blob/main/README.md#ai-agent-development)

## Prerequisites

* An [OCM offline token](https://console.redhat.com/openshift/token/rosa) for an account that can **read** (and, if you plan `apply`, **update**) the cluster in OCM
* The cluster’s **OCM cluster ID** (not only the display name). Obtain it from the Hybrid Cloud Console, or with the `ocm` CLI, for example: `ocm list clusters --columns=id,name`
* [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.0+
* A Terraform root module with the **osdgoogle** provider configured (see the [provider index](../index.md#authentication-and-configuration))

You do **not** need to declare WIF or VPC in this root module for import to succeed, as long as you only import `osdgoogle_cluster`.

## Step 1: Minimal configuration

Define a cluster resource with the **required** arguments for the [`osdgoogle_cluster` resource](../resources/cluster.md): `name`, `cloud_region`, and `gcp_project_id`. After import, Terraform **refresh** reads the cluster from OCM and fills many attributes from the API; starting values should match reality to reduce initial plan noise.

Example `main.tf`:

```terraform
terraform {
  required_providers {
    osdgoogle = {
      source  = "rh-mobb/osd-google"
      version = ">= 0.1.0"
    }
  }
}

provider "osdgoogle" {
  token = var.ocm_token
}

variable "ocm_token" {
  type        = string
  sensitive   = true
  description = "OCM offline token"
}

variable "cluster_name" {
  type        = string
  description = "Cluster name as shown in OCM (must match the imported cluster)"
}

variable "cloud_region" {
  type        = string
  description = "GCP region of the cluster (e.g. us-central1)"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID associated with the cluster"
}

resource "osdgoogle_cluster" "this" {
  name           = var.cluster_name
  cloud_region   = var.cloud_region
  gcp_project_id = var.gcp_project_id

  lifecycle {
    prevent_destroy = true
  }
}
```

Use `terraform.tfvars` or environment variables (`TF_VAR_*`) for non-secret values; pass the token via `-var 'ocm_token=...'` or a `*.tfvars` file that is **gitignored**.

Remove or relax `prevent_destroy` only after you trust `terraform plan` for this workspace.

## Step 2: Initialize

```console
terraform init
```

## Step 3: Import the cluster

Import ID is the **OCM cluster ID** (the value returned by OCM for the cluster resource).

```console
terraform import osdgoogle_cluster.this <OCM_CLUSTER_ID>
```

Terraform runs **Read** after import. If `name`, `cloud_region`, or `gcp_project_id` in configuration disagree with OCM, fix the variables and run **`terraform apply -refresh-only`** (or another plan/apply cycle) as appropriate so configuration matches refreshed state.

## Step 4: Plan and reconcile

Run:

```console
terraform plan
```

* The provider **updates** only a **subset** of cluster fields through OCM (for example **domain prefix**, **multi-AZ**, and **properties**). Other differences may appear as **configuration vs. state** until you align optional arguments or omit blocks you do not intend to manage.
* **Read** does not repopulate every optional block from OCM into state. Expect some **plan noise** on optional attributes until the configuration matches what you want Terraform to track.

Treat the first clean plan as validation—not a guarantee that every historical console choice is expressed in HCL.

## Optional: other osdgoogle resources

Other resources (for example **`osdgoogle_machine_pool`**, **`osdgoogle_wif_config`**, **`osdgoogle_cluster_admin`**) have their **own** import IDs and lifecycle. This guide does not document them end-to-end; see each [resource](../resources/) page and the [skill](https://github.com/rh-mobb/terraform-provider-osd-google/blob/main/.cursor/skills/osd-gcp-cluster-to-terraform/SKILL.md) reference for import formats and caveats.

## Summary

| Topic | This import-only guide | Broader adoption |
|-------|-------------------------|------------------|
| `osdgoogle_cluster` | Yes | Yes |
| `osdgoogle_wif_config` | No | Skill / custom docs |
| VPC / PSC / `google_*` | No | Skill / Google provider |
| Variableized multi-cluster roots | Minimal (your vars) | Skill workflow |

For anything beyond **binding one existing cluster** in OCM to **`osdgoogle_cluster`**, use the skill or extend your configuration using the [examples](https://github.com/rh-mobb/terraform-provider-osd-google/tree/main/examples) and [modules](https://github.com/rh-mobb/terraform-provider-osd-google/tree/main/modules) in this repository.
