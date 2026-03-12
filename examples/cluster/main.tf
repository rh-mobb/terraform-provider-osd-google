# OSD cluster with Workload Identity Federation (WIF)
#
# WIF allows OSD to assume GCP service account credentials without storing keys.
# WIF config is managed by terraform/wif_config/ (applied automatically by Makefile).
# This config looks up the WIF config by display_name and provisions GCP IAM + cluster.
#
# Prerequisites:
#   - OCM token (OSDGOOGLE_TOKEN or ocm_token variable)
#   - GCP project with WIF prerequisites (see OSD documentation)
#   - Application Default Credentials (gcloud auth application-default login)
#
# Usage: make example.cluster
# Both configs use cluster_name; display_name is "${cluster_name}-wif".

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

resource "osdgoogle_cluster" "cluster" {
  depends_on     = [module.wif_gcp]
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  version        = var.openshift_version
  wif_config_id  = data.osdgoogle_wif_config.wif.id
  compute_nodes  = 3
  ccs_enabled    = true

  lifecycle {
    prevent_destroy = false
  }
}

resource "osdgoogle_cluster_admin" "admin" {
  cluster_id = osdgoogle_cluster.cluster.id
  username   = "admin"
  password   = var.admin_password != "" ? var.admin_password : null
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
