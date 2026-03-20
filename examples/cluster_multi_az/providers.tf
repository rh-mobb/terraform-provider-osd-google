# Multi-AZ OSD cluster with bare metal machine pool (uses WIF)
#
# Creates a multi-AZ cluster with a VPC and a secondary machine pool using
# bare metal instance types. WIF config is managed by terraform/wif_config/.
#
# Prerequisites:
# - OCM token
# - GCP project with billing, OSD entitlements
# - Application Default Credentials (gcloud auth application-default login)

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
