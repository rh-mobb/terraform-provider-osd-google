# OSD cluster with module-managed VPC (uses WIF)
# WIF config managed by terraform/wif_config/. Cluster module handles WIF GCP + cluster.

module "osd_vpc" {
  source = "../../modules/osd-vpc"

  project_id             = var.gcp_project_id
  region                 = var.gcp_region
  cluster_name           = var.cluster_name
  enable_psc             = var.enable_psc
  enable_private_cluster = var.enable_psc # Enable firewall rules when PSC
}

module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  private_service_connect = var.enable_psc ? {
    service_attachment_subnet = module.osd_vpc.psc_subnet
  } : null

  security = var.enable_psc ? {
    secure_boot = true
  } : null

  machine_pools = var.machine_pools
}
