# OSD cluster with Shared VPC
#
# Uses a shared VPC (host project) for network connectivity instead
# of creating a new VPC in the service project.
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
  }
}

provider "osdgoogle" {
  token = var.ocm_token
}
