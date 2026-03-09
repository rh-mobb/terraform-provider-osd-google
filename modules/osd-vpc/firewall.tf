# Firewall rules for private OSD clusters
# Uses IP ranges (not tags) to match OSD-created instances

resource "google_compute_firewall" "cluster_internal" {
  count    = var.enable_private_cluster ? 1 : 0
  name     = "${var.cluster_name}-cluster-internal"
  network  = google_compute_network.vpc.id
  project  = var.project_id
  priority = 900

  allow {
    protocol = "all"
  }

  source_ranges = [
    var.master_cidr,
    var.worker_cidr,
  ]

  destination_ranges = [
    var.master_cidr,
    var.worker_cidr,
  ]

  direction = "INGRESS"
}

resource "google_compute_firewall" "psc_allow_https" {
  count    = var.enable_psc ? 1 : 0
  name     = "${var.cluster_name}-psc-allow-https"
  network  = google_compute_network.vpc.id
  project  = var.project_id
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }

  source_ranges = [
    var.master_cidr,
    var.worker_cidr,
  ]

  destination_ranges = [var.psc_cidr]
  direction          = "INGRESS"
}

resource "google_compute_firewall" "psc_allow_dns" {
  count    = var.enable_psc ? 1 : 0
  name     = "${var.cluster_name}-psc-allow-dns"
  network  = google_compute_network.vpc.id
  project  = var.project_id
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["53"]
  }

  allow {
    protocol = "udp"
    ports    = ["53"]
  }

  source_ranges = [
    var.master_cidr,
    var.worker_cidr,
  ]

  direction = "INGRESS"
}

resource "google_compute_firewall" "psc_internal_all" {
  count    = var.enable_psc ? 1 : 0
  name     = "${var.cluster_name}-psc-internal-all"
  network  = google_compute_network.vpc.id
  project  = var.project_id
  priority = 900

  allow {
    protocol = "all"
  }

  source_ranges = [
    var.master_cidr,
    var.worker_cidr,
    var.psc_cidr,
  ]

  direction = "INGRESS"
}

resource "google_compute_firewall" "bastion_to_cluster" {
  count    = var.enable_bastion_access ? 1 : 0
  name     = "${var.cluster_name}-bastion-to-cluster"
  network  = google_compute_network.vpc.id
  project  = var.project_id
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["6443", "22623", "443", "80"]
  }

  source_ranges = [var.bastion_cidr]

  destination_ranges = [
    var.master_cidr,
    var.worker_cidr,
  ]

  direction = "INGRESS"
}
