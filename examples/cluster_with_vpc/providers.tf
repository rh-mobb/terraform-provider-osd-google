# OSD cluster with module-managed VPC (uses WIF)
#
# Creates a custom VPC with control plane and worker subnets, then
# deploys an OSD cluster. WIF config is managed by terraform/wif_config/.
# Optionally enables PSC for private clusters.
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
