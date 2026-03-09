terraform {
  required_providers {
    # Use terraform.local for local dev (make install). When published, use registry.terraform.io/rh-mobb/osd-google
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
  token             = var.ocm_token != "" ? var.ocm_token : null
  openshift_version = var.openshift_version
}

provider "google" {
  project = var.gcp_project_id
}
