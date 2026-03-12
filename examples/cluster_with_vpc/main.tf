# OSD cluster with module-managed VPC (uses WIF)
# WIF config managed by terraform/wif_config/. Uses data source + wif_gcp module.

data "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
}

data "google_project" "project" {
  project_id = var.gcp_project_id
}

module "wif_gcp" {
  source = "../../modules/osd-wif-gcp"

  project_id   = var.gcp_project_id
  display_name = data.osdgoogle_wif_config.wif.display_name
  pool_id      = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.pool_id
  identity_provider = {
    identity_provider_id = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.identity_provider_id
    issuer_url           = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.issuer_url
    jwks                 = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.jwks
    allowed_audiences    = data.osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.allowed_audiences
  }
  service_accounts         = data.osdgoogle_wif_config.wif.gcp.service_accounts
  support                  = data.osdgoogle_wif_config.wif.gcp.support
  impersonator_email       = data.osdgoogle_wif_config.wif.gcp.impersonator_email
  federated_project_id     = try(data.osdgoogle_wif_config.wif.gcp.federated_project_id, null) != "" ? try(data.osdgoogle_wif_config.wif.gcp.federated_project_id, null) : null
  federated_project_number = try(data.osdgoogle_wif_config.wif.gcp.federated_project_number, "") != "" ? data.osdgoogle_wif_config.wif.gcp.federated_project_number : tostring(data.google_project.project.number)
}

module "osd_vpc" {
  source = "../../modules/osd-vpc"

  project_id             = var.gcp_project_id
  region                 = var.gcp_region
  cluster_name           = var.cluster_name
  enable_psc             = var.enable_psc
  enable_private_cluster = var.enable_psc # Enable firewall rules when PSC
}

resource "osdgoogle_cluster" "cluster" {
  depends_on = [module.wif_gcp]

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id
  wif_config_id  = data.osdgoogle_wif_config.wif.id
  version        = var.openshift_version
  compute_nodes  = 3
  ccs_enabled    = true

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
}

locals {
  machine_pools_map = { for mp in var.machine_pools : mp.name => mp }
}

resource "osdgoogle_machine_pool" "pools" {
  for_each = local.machine_pools_map

  cluster_id    = osdgoogle_cluster.cluster.id
  name          = each.value.name
  instance_type = each.value.instance_type

  replicas = each.value.autoscaling_enabled ? null : each.value.replicas
  autoscaling = each.value.autoscaling_enabled ? {
    min_replicas = each.value.min_replicas
    max_replicas = each.value.max_replicas
  } : null

  availability_zones = each.value.availability_zones
  labels             = each.value.labels
  taints             = each.value.taints
  root_volume_size   = each.value.root_volume_size

  gcp = try(each.value.secure_boot, false) ? { secure_boot = true } : null
}
