# Basic OSD cluster on GCP
#
# Prerequisites:
# - OCM token (https://console.redhat.com/openshift/token/rosa)
# - GCP project with billing enabled
# - OCM + GCP account linking for OSD
#
# Usage:
#   export OCM_TOKEN="your-token"
#   terraform init
#   terraform plan
#   terraform apply

terraform {
  required_providers {
    osdgoogle = {
      source  = "registry.terraform.io/redhat/osd-google"
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

variable "cluster_name" {
  type        = string
  default     = "my-osd-cluster"
  description = "Name of the cluster"
}

resource "osdgoogle_cluster" "example" {
  name                  = var.cluster_name
  cloud_region          = "us-central1"
  gcp_project_id        = var.gcp_project_id
  version               = "4.16.1"
  compute_nodes         = 3
  compute_machine_type  = "custom-4-16384"
  ccs_enabled           = true
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
