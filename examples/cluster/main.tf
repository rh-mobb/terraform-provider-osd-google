# OSD cluster with Workload Identity Federation (WIF)
#
# WIF config is managed by terraform/wif_config/ (applied automatically by Makefile).
# Usage: make example.cluster

module "cluster" {
  source = "../../modules/osd-cluster"

  name              = var.cluster_name
  cloud_region      = "us-central1"
  gcp_project_id    = var.gcp_project_id
  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  create_admin   = true
  admin_password = var.admin_password != "" ? var.admin_password : null

  machine_pools = var.machine_pools
}
