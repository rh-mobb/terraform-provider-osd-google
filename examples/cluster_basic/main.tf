# Basic OSD cluster on GCP (CCS with service account)
#
# CCS cluster using GCP service account credentials (no WIF). Assumes
# osd-ccs-admin already exists; looks it up and creates a key for it.
# For WIF-based CCS, see cluster_wif.
#
# Prerequisites (no networks or subnets to pre-create; OSD creates them):
# - OCM token (https://console.redhat.com/openshift/token/rosa)
# - GCP project with billing, OSD entitlements, and required APIs enabled
# - Pre-created osd-ccs-admin service account with OSD roles (see
#   https://docs.openshift.com/dedicated/osd_planning/gcp-ccs.html)
# - Application Default Credentials (gcloud auth application-default login)
#   or GOOGLE_APPLICATION_CREDENTIALS for the Google provider
#
# Local development:
#   make install                  # Install provider to ~/.terraform.d/plugins
#   export OSDGOOGLE_TOKEN="..."   # Or OSDGOOGLE_CLIENT_ID + OSDGOOGLE_CLIENT_SECRET
#   gcloud auth application-default login
#   terraform init
#   terraform plan -var="gcp_project_id=YOUR_PROJECT"
#   terraform apply -var="gcp_project_id=YOUR_PROJECT"
#
# Production (when published): change source to registry.terraform.io/redhat/osd-google

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
  # Token from OSDGOOGLE_TOKEN env var when not set here
}

provider "google" {
  project = var.gcp_project_id
}

variable "ocm_token" {
  type        = string
  sensitive   = true
  description = "OCM token (optional; set token = var.ocm_token in provider to use)"
  default     = ""
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID for the cluster"
}

variable "cluster_name" {
  type        = string
  default     = "my-osd-cluster"
  description = "Name of the cluster"
}

data "google_service_account" "osd_ccs_admin" {
  account_id = "osd-ccs-admin"
}

resource "google_service_account_key" "osd_ccs" {
  service_account_id = data.google_service_account.osd_ccs_admin.name
}

locals {
  sa_key_json = jsondecode(base64decode(google_service_account_key.osd_ccs.private_key))
}

resource "osdgoogle_cluster" "example" {
  name                 = var.cluster_name
  cloud_region          = "us-central1"
  gcp_project_id        = var.gcp_project_id
  # version               = "4.16.1"
  compute_nodes         = 3
  compute_machine_type  = "custom-4-16384"
  ccs_enabled           = true

  gcp_authentication = {
    client_email   = local.sa_key_json.client_email
    client_id      = local.sa_key_json.client_id
    private_key_id = local.sa_key_json.private_key_id
    private_key    = local.sa_key_json.private_key
  }
}

output "osd_ccs_admin_email" {
  value       = data.google_service_account.osd_ccs_admin.email
  description = "OSD CCS admin service account email"
}

output "cluster_id" {
  value       = osdgoogle_cluster.example.id
  description = "OCM cluster ID"
}

output "cluster_state" {
  value       = osdgoogle_cluster.example.state
  description = "Cluster state"
}

output "api_url" {
  value       = osdgoogle_cluster.example.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = osdgoogle_cluster.example.console_url
  description = "OpenShift web console URL"
}
