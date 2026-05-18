# Private OSD cluster with IAP bastion for SSH tunneling
#
# Creates a fully private cluster (API endpoint internal-only) behind a
# module-managed VPC with Private Service Connect. A CentOS bastion VM
# with no external IP is deployed in the same VPC and reachable via
# gcloud IAP SSH tunneling. WIF config is managed by terraform/wif_config/.
#
# Prerequisites:
# - OCM token (OSDGOOGLE_TOKEN env var or ocm_token variable)
# - GCP project with billing, OSD entitlements, and IAP API enabled
# - Application Default Credentials: gcloud auth application-default login
# - gcloud CLI for IAP SSH access (see make example.cluster_private.ssh)

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
