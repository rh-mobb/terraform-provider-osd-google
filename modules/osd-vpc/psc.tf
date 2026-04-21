# Private Service Connect resources (optional)
# Required for private OSD clusters with PSC

locals {
  # GCP PSC Google APIs forwarding rules allow only lowercase letters and digits (no hyphens),
  # max 20 characters. Strip hyphens and truncate the cluster name prefix to fit.
  psc_rule_name = "${substr(replace(var.cluster_name, "-", ""), 0, 13)}pscapis"
}

resource "google_compute_subnetwork" "psc" {
  count         = var.enable_psc ? 1 : 0
  name          = "${var.cluster_name}-psc-subnet"
  ip_cidr_range = var.psc_cidr
  region        = var.region
  network       = google_compute_network.vpc.id
  purpose       = "PRIVATE_SERVICE_CONNECT"
  project       = var.project_id
}

resource "google_compute_global_address" "psc" {
  count        = var.enable_psc ? 1 : 0
  name         = "${var.cluster_name}-psc-ip"
  purpose      = "PRIVATE_SERVICE_CONNECT"
  address_type = "INTERNAL"
  address      = "10.0.255.100" # Outside all subnets
  network      = google_compute_network.vpc.id
  project      = var.project_id
}

resource "google_compute_global_forwarding_rule" "psc" {
  count                 = var.enable_psc ? 1 : 0
  name                  = local.psc_rule_name
  target                = "all-apis"
  network               = google_compute_network.vpc.id
  ip_address            = google_compute_global_address.psc[0].id
  load_balancing_scheme = ""
  project               = var.project_id
}

resource "google_dns_managed_zone" "psc_googleapis" {
  count       = var.enable_psc ? 1 : 0
  name        = "${var.cluster_name}-googleapis"
  dns_name    = "googleapis.com."
  description = "Private DNS zone for Google APIs via PSC"
  project     = var.project_id

  visibility = "private"

  private_visibility_config {
    networks {
      network_url = google_compute_network.vpc.id
    }
  }
}

resource "google_dns_record_set" "psc_googleapis_a" {
  count        = var.enable_psc ? 1 : 0
  name         = "*.${google_dns_managed_zone.psc_googleapis[0].dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = google_dns_managed_zone.psc_googleapis[0].name
  rrdatas      = [google_compute_global_address.psc[0].address]
  project      = var.project_id
}

resource "google_dns_managed_zone" "psc_gcr" {
  count       = var.enable_psc ? 1 : 0
  name        = "${var.cluster_name}-gcr"
  dns_name    = "gcr.io."
  description = "Private DNS zone for GCR via PSC"
  project     = var.project_id

  visibility = "private"

  private_visibility_config {
    networks {
      network_url = google_compute_network.vpc.id
    }
  }
}

resource "google_dns_record_set" "psc_gcr_a" {
  count        = var.enable_psc ? 1 : 0
  name         = "*.${google_dns_managed_zone.psc_gcr[0].dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = google_dns_managed_zone.psc_gcr[0].name
  rrdatas      = [google_compute_global_address.psc[0].address]
  project      = var.project_id
}
