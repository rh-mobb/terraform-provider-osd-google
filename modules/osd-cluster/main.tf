# OSD cluster module: WIF GCP provisioning + cluster + admin + machine pools
#
# WIF config must exist in OCM before applying (e.g. via terraform/wif_config/).
# This module looks up the WIF config by display_name, provisions GCP IAM, then creates the cluster.

locals {
  wif_config_display_name = coalesce(var.wif_config_display_name, "${var.name}-wif")
}

data "osdgoogle_wif_config" "wif" {
  display_name = local.wif_config_display_name
}

data "google_project" "project" {
  project_id = var.gcp_project_id
}

module "wif_gcp" {
  source = "../osd-wif-gcp"

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

resource "osdgoogle_cluster" "cluster" {
  depends_on = [module.wif_gcp]

  name           = var.name
  cloud_region   = var.cloud_region
  gcp_project_id = var.gcp_project_id
  wif_config_id  = data.osdgoogle_wif_config.wif.id
  version        = var.openshift_version
  compute_nodes  = var.compute_nodes
  ccs_enabled    = var.ccs_enabled

  multi_az                 = var.multi_az
  availability_zones       = var.availability_zones
  domain_prefix            = var.domain_prefix
  billing_model            = var.billing_model
  properties               = var.properties
  compute_machine_type     = var.compute_machine_type
  gcp_network              = var.gcp_network
  private_service_connect  = var.private_service_connect
  security                 = var.security
  network                  = var.network
  autoscaling              = var.autoscaling
  wait_for_create_complete = var.wait_for_create_complete
  wait_timeout             = var.wait_timeout
}

resource "osdgoogle_cluster_admin" "admin" {
  count = var.create_admin ? 1 : 0

  cluster_id = osdgoogle_cluster.cluster.id
  username   = var.admin_username
  password   = var.admin_password
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
