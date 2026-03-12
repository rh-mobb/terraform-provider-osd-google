# WIF config for OSD clusters
#
# Creates the WIF config in OCM. Applied automatically by the Makefile
# before any example (make example.<name>).
#
# Prerequisites:
#   - OCM token (OSDGOOGLE_TOKEN or ocm_token variable)
#   - GCP project with WIF prerequisites (see OSD documentation)
#   - Application Default Credentials (gcloud auth application-default login)

data "google_project" "project" {
  project_id = var.gcp_project_id
}

resource "osdgoogle_wif_config" "wif" {
  display_name      = "${var.cluster_name}-wif"
  openshift_version = var.openshift_version
  gcp = {
    project_id     = var.gcp_project_id
    project_number = tostring(data.google_project.project.number)
    role_prefix    = replace(replace(coalesce(var.role_prefix, var.cluster_name), "-", ""), "_", "")
  }
}
