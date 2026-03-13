# OSD cluster with Private Service Connect (PSC) and Secure Boot
# WIF config managed by terraform/wif_config/. Cluster module handles WIF GCP + cluster.

module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  private_service_connect = {
    service_attachment_subnet = var.psc_subnet
  }

  security = {
    secure_boot = true
  }

  machine_pools = var.machine_pools
}
