resource "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
  gcp = {
    project_id     = var.gcp_project_id
    project_number = var.gcp_project_number
    role_prefix    = "osd"
  }
}

resource "osdgoogle_cluster" "psc_cluster" {
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  wif_config_id  = osdgoogle_wif_config.wif.id
  version        = "4.16.1"
  compute_nodes  = 3
  ccs_enabled    = true

  private_service_connect = {
    service_attachment_subnet = var.psc_subnet
  }

  security = {
    secure_boot = true
  }
}
