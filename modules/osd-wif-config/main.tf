# OCM Workload Identity Federation (WIF) config for OSD clusters on GCP
# Creates the WIF config in OCM. Used by terraform/wif_config/ as Phase 1 of the two-phase apply workflow.

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
