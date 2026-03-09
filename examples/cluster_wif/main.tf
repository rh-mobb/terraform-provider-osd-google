# OSD cluster with Workload Identity Federation (WIF)
#
# WIF allows OSD to assume GCP service account credentials without storing keys.
# Create the WIF config in OCM first; OCM returns a blueprint (pool, SAs, IAM).
# The GCP module then provisions the pool, service accounts, and bindings.
#
# Prerequisites:
#   - OCM token (OSDGOOGLE_TOKEN or ocm_token variable)
#   - GCP project with WIF prerequisites (see OSD documentation)
#   - Application Default Credentials (gcloud auth application-default login)
#
# Two-phase apply required: The module's for_each (service_accounts, support)
# depends on OCM's blueprint, which Terraform only knows after wif_config exists.
# Use:  make apply-wif-cluster
# Or:   terraform apply -target=osdgoogle_wif_config.wif && terraform apply
# See:  examples/cluster_wif/README.md

data "google_project" "project" {
  project_id = var.gcp_project_id
}

resource "osdgoogle_wif_config" "wif" {
  display_name       = "${var.cluster_name}-wif"
  openshift_version  = var.openshift_version
  gcp = {
    project_id     = var.gcp_project_id
    project_number = tostring(data.google_project.project.number)
    role_prefix    = replace(replace(coalesce(var.role_prefix, var.cluster_name), "-", ""), "_", "")
  }
}

module "wif_gcp" {
  source = "../../modules/osd-wif-gcp"

  project_id         = var.gcp_project_id
  display_name       = osdgoogle_wif_config.wif.display_name
  pool_id            = osdgoogle_wif_config.wif.gcp.workload_identity_pool.pool_id
  identity_provider = {
    identity_provider_id = osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.identity_provider_id
    issuer_url           = osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.issuer_url
    jwks                 = osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.jwks
    allowed_audiences    = osdgoogle_wif_config.wif.gcp.workload_identity_pool.identity_provider.allowed_audiences
  }
  service_accounts         = osdgoogle_wif_config.wif.gcp.service_accounts
  support                  = osdgoogle_wif_config.wif.gcp.support
  impersonator_email       = osdgoogle_wif_config.wif.gcp.impersonator_email
  federated_project_id     = try(osdgoogle_wif_config.wif.gcp.federated_project_id, null) != "" ? try(osdgoogle_wif_config.wif.gcp.federated_project_id, null) : null
  federated_project_number = try(osdgoogle_wif_config.wif.gcp.federated_project_number, "") != "" ? osdgoogle_wif_config.wif.gcp.federated_project_number : tostring(data.google_project.project.number)
}

resource "osdgoogle_cluster" "wif_cluster" {
  depends_on     = [module.wif_gcp]
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  version        = var.openshift_version
  wif_config_id  = osdgoogle_wif_config.wif.id
  compute_nodes  = 3
  ccs_enabled    = true

  lifecycle {
    prevent_destroy = false
  }
}

resource "osdgoogle_cluster_admin" "admin" {
  cluster_id = osdgoogle_cluster.wif_cluster.id
  username   = "admin"
  password   = var.admin_password != "" ? var.admin_password : null
}
