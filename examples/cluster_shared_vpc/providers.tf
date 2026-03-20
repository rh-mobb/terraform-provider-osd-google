# OSD cluster with Shared VPC (uses WIF)
#
# Uses a shared VPC (host project) for network connectivity.
# WIF config managed by terraform/wif_config/. Uses data source + wif_gcp module.
#
# Prerequisites:
# - OCM token
# - Shared VPC host project
# - Service project attached to the shared VPC
# - Subnets for control plane and compute in the host project

terraform {
  required_providers {
    osdgoogle = {
      source  = "registry.terraform.io/rh-mobb/osd-google"
      version = ">= 0.0.1"
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
