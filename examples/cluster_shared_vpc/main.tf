# OSD cluster with Shared VPC
#
# Uses a shared VPC (host project) for network connectivity instead
# of creating a new VPC in the service project.
#
# Prerequisites:
# - OCM token
# - Shared VPC host project
# - Service project attached to the shared VPC
# - Subnets for control plane and compute in the host project

terraform {
  required_providers {
    osdgoogle = {
      source  = "terraform.local/local/osd-google"  # use registry.terraform.io/redhat/osd-google when published
      version = ">= 0.0.1"
    }
  }
}

provider "osdgoogle" {
  token = var.ocm_token
}

variable "ocm_token" {
  type        = string
  sensitive   = true
  description = "OCM offline token or access token"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP service project ID (where cluster resources are created)"
}

variable "gcp_project_number" {
  type        = string
  description = "GCP project number (numeric)"
}

variable "vpc_host_project_id" {
  type        = string
  description = "GCP host project ID (shared VPC owner)"
}

variable "vpc_name" {
  type        = string
  description = "Name of the shared VPC network"
}

variable "compute_subnet" {
  type        = string
  description = "Subnet name for worker nodes"
}

variable "control_plane_subnet" {
  type        = string
  description = "Subnet name for control plane nodes"
}

variable "cluster_name" {
  type        = string
  default     = "my-shared-vpc-cluster"
  description = "Name of the cluster"
}

resource "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
  gcp {
    project_id     = var.gcp_project_id
    project_number = var.gcp_project_number
    role_prefix    = "osd"
  }
}

resource "osdgoogle_cluster" "shared_vpc_cluster" {
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  wif_config_id  = osdgoogle_wif_config.wif.id
  version        = "4.16.1"
  compute_nodes  = 3
  ccs_enabled    = true

  gcp_network = {
    vpc_name             = var.vpc_name
    vpc_project_id       = var.vpc_host_project_id
    compute_subnet       = var.compute_subnet
    control_plane_subnet = var.control_plane_subnet
  }
}

output "cluster_id" {
  value = osdgoogle_cluster.shared_vpc_cluster.id
}
