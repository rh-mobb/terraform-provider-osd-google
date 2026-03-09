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
      source  = "registry.terraform.io/rh-mobb/osd-google"
      version = ">= 0.0.1"
    }
  }
}

provider "osdgoogle" {
  token = var.ocm_token
}
