# OSD cluster with cluster admin user
#
# Creates an OSD cluster and adds an HTPasswd-based cluster admin user
# with cluster-admins group membership. The admin can log in via
# `oc login` using the identity provider credentials.
#
# Prerequisites: same as cluster_basic (OCM token, GCP project, osd-ccs-admin SA)
#
# Usage:
#   make dev-setup                # Build binary and print ~/.terraformrc config
#   export OSDGOOGLE_TOKEN="..."
#   gcloud auth application-default login
#   terraform plan -var="gcp_project_id=YOUR_PROJECT"
#   terraform apply -var="gcp_project_id=YOUR_PROJECT"

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

provider "osdgoogle" {}

provider "google" {
  project = var.gcp_project_id
}
