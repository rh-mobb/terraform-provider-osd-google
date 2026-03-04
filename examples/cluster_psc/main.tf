# OSD cluster with Private Service Connect (PSC)
#
# PSC allows private connectivity to OCM/Red Hat services without
# exposing traffic to the public internet.
#
# Prerequisites:
# - OCM token
# - GCP project with a subnet reserved for PSC service attachments

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
  description = "GCP project ID for the cluster"
}

variable "gcp_project_number" {
  type        = string
  description = "GCP project number (numeric)"
}

variable "psc_subnet" {
  type        = string
  description = "Subnet name or CIDR for PSC service attachment"
}

variable "cluster_name" {
  type        = string
  default     = "my-psc-cluster"
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

resource "osdgoogle_cluster" "psc_cluster" {
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  wif_config_id  = osdgoogle_wif_config.wif.id
  version        = "4.16.1"
  compute_nodes  = 3
  ccs_enabled    = true

  private_service_connect = {
    service_attachment_subnet = var.psc_subnet
  }

  security = {
    secure_boot = true
  }
}

output "cluster_id" {
  value = osdgoogle_cluster.psc_cluster.id
}
