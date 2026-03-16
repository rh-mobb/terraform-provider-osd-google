# WIF config for OSD clusters
#
# Creates the WIF config in OCM. Applied automatically by the Makefile
# before any example (make example.<name>).
#
# Prerequisites:
#   - OCM token (OSDGOOGLE_TOKEN or ocm_token variable)
#   - GCP project with WIF prerequisites (see OSD documentation)
#   - Application Default Credentials (gcloud auth application-default login)

module "wif_config" {
  source = "../../modules/osd-wif-config"

  gcp_project_id    = var.gcp_project_id
  cluster_name      = var.cluster_name
  openshift_version = var.openshift_version
  role_prefix       = var.role_prefix
}
