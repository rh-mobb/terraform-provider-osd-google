# Multi-AZ OSD cluster
#
# Uses cluster module (WIF config created by terraform/wif_config/).
# The cluster is deployed across multiple availability zones with module-managed VPC.
# Explicitly passes 3 zones (required for multi-AZ when specifying availability_zones).

data "google_compute_zones" "available" {
  project = var.gcp_project_id
  region  = var.gcp_region
  status  = "UP"
}

module "osd_vpc" {
  source = "../../modules/osd-vpc"

  project_id   = var.gcp_project_id
  region       = var.gcp_region
  cluster_name = var.cluster_name
}

module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  openshift_version  = var.openshift_version
  multi_az           = true
  compute_nodes      = 3
  availability_zones = slice(data.google_compute_zones.available.names, 0, 3)
  ccs_enabled        = true

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  machine_pools = var.machine_pools
}
