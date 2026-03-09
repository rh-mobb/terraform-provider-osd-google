resource "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
  gcp = {
    project_id     = var.gcp_project_id
    project_number = var.gcp_project_number
    role_prefix    = "osd"
  }
}

resource "osdgoogle_cluster" "shared_vpc_cluster" {
  name           = var.cluster_name
  cloud_region   = "us-central1"
  gcp_project_id = var.gcp_project_id
  wif_config_id  = osdgoogle_wif_config.wif.id
  version        = "4.16.1"
  compute_nodes  = 3
  ccs_enabled    = true

  gcp_network = {
    vpc_name             = var.vpc_name
    vpc_project_id       = var.vpc_host_project_id
    compute_subnet       = var.compute_subnet
    control_plane_subnet = var.control_plane_subnet
  }
}
