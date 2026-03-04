# OSD cluster with Workload Identity Federation (WIF)
#
# WIF allows OSD to assume GCP service account credentials without
# storing keys. Create the WIF config first, then reference it in the cluster.
#
# Prerequisites:
# - OCM token
# - GCP project with WIF prerequisites (see OSD documentation)
# - Application Default Credentials (gcloud auth application-default login)
#   for the Google provider to look up project details

terraform {
  required_providers {
    osdgoogle = {
      source  = "terraform.local/local/osd-google"
      version = ">= 0.0.1"
    }
    google = {
      source  = "hashicorp/google"
      version = ">= 5.0"
    }
  }
}

provider "osdgoogle" {
  token = var.ocm_token
}

provider "google" {
  project = var.gcp_project_id
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

variable "role_prefix" {
  type        = string
  default     = "osd"
  description = "Prefix for custom IAM roles in GCP"
}

variable "cluster_name" {
  type        = string
  default     = "my-wif-cluster"
  description = "Name of the cluster"
}

data "google_project" "project" {
  project_id = var.gcp_project_id
}

resource "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
  gcp = {
    project_id     = var.gcp_project_id
    project_number = tostring(data.google_project.project.number)
    role_prefix    = var.role_prefix
  }
}

resource "osdgoogle_cluster" "wif_cluster" {
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  # version        = "4.16.1"
  wif_config_id  = osdgoogle_wif_config.wif.id
  compute_nodes  = 3
  ccs_enabled    = true
}

output "cluster_id" {
  value = osdgoogle_cluster.wif_cluster.id
}

output "api_url" {
  value = osdgoogle_cluster.wif_cluster.api_url
}
