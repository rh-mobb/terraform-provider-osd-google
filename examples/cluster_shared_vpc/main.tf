# OSD cluster using Shared VPC
# WIF config managed by terraform/wif_config/. Cluster module handles WIF GCP + cluster.

module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  gcp_network = {
    vpc_name             = var.vpc_name
    vpc_project_id       = var.vpc_host_project_id
    compute_subnet       = var.compute_subnet
    control_plane_subnet = var.control_plane_subnet
  }

  machine_pools = var.machine_pools
}
