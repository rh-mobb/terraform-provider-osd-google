# OSD cluster with bare metal as default compute instance type (single AZ)
#
# Uses bare metal (e.g. c3-standard-192-metal) for the default worker nodes.
# Single-AZ cluster. OCM supports multi-zone for bare metal, but the machine
# type must be available in each zone; c3-standard-192-metal is only in
# us-central1-a (not us-central1-b). Specify zones where the type exists.
#
# NOTE: Secure Boot (Shielded VMs) is NOT supported on bare metal instance types.
#
# Prerequisites:
#   - OCM token (OSDGOOGLE_TOKEN or ocm_token variable)
#   - GCP project with WIF prerequisites (see OSD documentation)
#   - Application Default Credentials (gcloud auth application-default login)
#
# Usage: make example.cluster_baremetal
# Both configs use cluster_name; display_name is "${cluster_name}-wif".

data "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
}

data "osdgoogle_machine_types" "baremetal" {
  region         = var.gcp_region
  gcp_project_id = var.gcp_project_id
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

locals {
  machine_type_ids   = [for item in data.osdgoogle_machine_types.baremetal.items : item.id]
  machine_type_valid = contains(local.machine_type_ids, var.compute_machine_type)
}

resource "osdgoogle_cluster" "cluster" {
  depends_on = [module.wif_gcp]

  name                 = var.cluster_name
  cloud_region         = var.gcp_region
  gcp_project_id       = var.gcp_project_id
  version              = var.openshift_version
  wif_config_id        = data.osdgoogle_wif_config.wif.id
  compute_nodes        = var.compute_nodes
  compute_machine_type = var.compute_machine_type
  availability_zones   = [var.availability_zone]
  ccs_enabled          = true

  lifecycle {
    precondition {
      condition     = local.machine_type_valid
      error_message = <<-EOT
        Instance type '${var.compute_machine_type}' is not available in region '${var.gcp_region}'.
        Available types: ${join(", ", local.machine_type_ids)}
        For bare metal, specify availability_zones with zones where the machine type exists (e.g. us-central1-a; us-central1-b does NOT support c3-standard-192-metal).
      EOT
    }
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
