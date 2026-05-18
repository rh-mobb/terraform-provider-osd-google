# VPC network for OSD clusters (BYOVPC)
# Based on terraform-google-osd reference

resource "google_compute_network" "vpc" {
  project                 = var.project_id
  name                    = "${var.cluster_name}-vpc"
  auto_create_subnetworks = false
  routing_mode            = var.routing_mode
}

resource "google_compute_subnetwork" "master" {
  project                  = var.project_id
  name                     = "${var.cluster_name}-master-subnet"
  ip_cidr_range            = var.master_cidr
  region                   = var.region
  network                  = google_compute_network.vpc.id
  private_ip_google_access = true
}

resource "google_compute_subnetwork" "worker" {
  project                  = var.project_id
  name                     = "${var.cluster_name}-worker-subnet"
  ip_cidr_range            = var.worker_cidr
  region                   = var.region
  network                  = google_compute_network.vpc.id
  private_ip_google_access = true
}

resource "google_compute_router" "router" {
  project = var.project_id
  name    = "${var.cluster_name}-router"
  region  = var.region
  network = google_compute_network.vpc.id
}

resource "google_compute_router_nat" "nat_master" {
  name                               = "${var.cluster_name}-nat-master"
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
  subnetwork {
    name                    = google_compute_subnetwork.master.id
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
  min_ports_per_vm                    = 7168
  enable_endpoint_independent_mapping = false
}

resource "google_compute_router_nat" "nat_worker" {
  name                               = "${var.cluster_name}-nat-worker"
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
  subnetwork {
    name                    = google_compute_subnetwork.worker.id
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
  min_ports_per_vm                    = 4096
  enable_endpoint_independent_mapping = false
}
