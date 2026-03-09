data "google_project" "project" {
  project_id = var.gcp_project_id
}

module "osd_vpc" {
  source = "../../modules/osd-vpc"

  project_id            = var.gcp_project_id
  region                = var.gcp_region
  cluster_name          = var.cluster_name
  enable_psc            = var.enable_psc
  enable_private_cluster = var.enable_psc # Enable firewall rules when PSC
}

resource "osdgoogle_wif_config" "wif" {
  display_name = "${var.cluster_name}-wif"
  gcp = {
    project_id     = var.gcp_project_id
    project_number = tostring(data.google_project.project.number)
    role_prefix    = "osd"
  }
}

resource "osdgoogle_cluster" "cluster" {
  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id
  wif_config_id  = osdgoogle_wif_config.wif.id
  version        = "4.16.1"
  compute_nodes  = 3
  ccs_enabled    = true

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    vpc_project_id       = var.gcp_project_id
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
