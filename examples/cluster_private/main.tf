# Private OSD cluster with IAP bastion
#
# Network layout (all CIDRs within the module-managed VPC):
#   10.0.0.0/19   - control plane (master) subnet
#   10.0.32.0/19  - worker/compute subnet
#   10.0.64.0/29  - Private Service Connect subnet
#   10.0.128.0/24 - dedicated bastion subnet (when bastion_use_worker_subnet = false)
#
# Set bastion_use_worker_subnet = true to place the bastion in the worker subnet,
# giving it identical firewall rules, NAT, and Private Google Access to worker nodes.
#
# Access pattern:
#   gcloud compute ssh <bastion> --tunnel-through-iap   (interactive shell)
#   make example.cluster_private.ssh                    (convenience target)
#   make example.cluster_private.tunnel                 (port-forward API to localhost:6443)

locals {
  # When using the worker subnet the bastion inherits the worker firewall rules and NAT.
  # When using a dedicated subnet the bastion_to_cluster firewall rule (via enable_bastion_access)
  # allows it to reach the cluster; a separate NAT is also created below.
  bastion_subnetwork = (
    var.bastion_use_worker_subnet
    ? module.osd_vpc.compute_subnet
    : google_compute_subnetwork.bastion[0].self_link
  )
}

# --------------------------------------------------------------------------
# VPC: subnets, Cloud NAT, PSC, private-cluster firewall rules
# --------------------------------------------------------------------------

module "osd_vpc" {
  source = "../../modules/osd-vpc"

  project_id             = var.gcp_project_id
  region                 = var.gcp_region
  cluster_name           = var.cluster_name
  enable_psc             = true
  enable_private_cluster = true
  # bastion_to_cluster firewall rule only needed when bastion has its own subnet
  enable_bastion_access = !var.bastion_use_worker_subnet
  bastion_cidr          = "10.0.128.0/24"
}

# --------------------------------------------------------------------------
# Dedicated bastion subnet (skipped when bastion_use_worker_subnet = true)
# --------------------------------------------------------------------------

resource "google_compute_subnetwork" "bastion" {
  count         = var.bastion_use_worker_subnet ? 0 : 1
  name          = "${var.cluster_name}-bastion-subnet"
  ip_cidr_range = "10.0.128.0/24"
  region        = var.gcp_region
  network       = module.osd_vpc.vpc_id
  project       = var.gcp_project_id
}

# Cloud NAT for the dedicated bastion subnet (skipped when bastion_use_worker_subnet = true
# because the worker subnet's NAT already covers the bastion in that case).
resource "google_compute_router_nat" "nat_bastion" {
  count                              = var.bastion_use_worker_subnet ? 0 : 1
  name                               = "${var.cluster_name}-nat-bastion"
  router                             = "${var.cluster_name}-router"
  region                             = var.gcp_region
  project                            = var.gcp_project_id
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"

  subnetwork {
    name                    = google_compute_subnetwork.bastion[0].id
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }

  depends_on = [module.osd_vpc]
}

# --------------------------------------------------------------------------
# Firewall: allow Google IAP range to reach bastion on SSH (port 22)
# Docs: https://cloud.google.com/iap/docs/using-tcp-forwarding
# --------------------------------------------------------------------------

resource "google_compute_firewall" "iap_ssh" {
  name    = "${var.cluster_name}-iap-ssh"
  network = module.osd_vpc.vpc_id
  project = var.gcp_project_id

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  # Google's IAP TCP forwarding source range (static, documented by Google)
  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["iap-ssh"]
}

# --------------------------------------------------------------------------
# Bastion VM: CentOS Stream 9, no external IP, OS Login enabled
# Reached exclusively via IAP SSH tunnel (gcloud compute ssh --tunnel-through-iap)
# --------------------------------------------------------------------------

data "google_compute_image" "centos" {
  family  = "centos-stream-9"
  project = "centos-cloud"
}

resource "google_compute_instance" "bastion" {
  name         = "${var.cluster_name}-bastion"
  machine_type = var.bastion_machine_type
  zone         = var.bastion_zone
  project      = var.gcp_project_id

  tags = ["iap-ssh"]

  boot_disk {
    initialize_params {
      image = data.google_compute_image.centos.self_link
      size  = 20
      type  = "pd-standard"
    }
  }

  network_interface {
    subnetwork = local.bastion_subnetwork
    # No access_config block = no external IP
  }

  metadata = {
    # OS Login lets gcloud manage SSH keys instead of project-wide metadata
    enable-oslogin = "TRUE"
  }

  # Install oc/kubectl on first boot so the bastion is ready to use.
  # $${...} is Terraform's escape for a literal ${...} in bash.
  metadata_startup_script = <<-EOT
    #!/bin/bash
    set -euo pipefail
    OC_MIRROR="https://mirror.openshift.com/pub/openshift-v4/clients/ocp/stable"
    curl -sL "$${OC_MIRROR}/openshift-client-linux.tar.gz" \
      | tar -xz -C /usr/local/bin oc kubectl
    chmod +x /usr/local/bin/oc /usr/local/bin/kubectl
  EOT

  shielded_instance_config {
    enable_secure_boot = true
  }
}

# --------------------------------------------------------------------------
# Private OSD cluster
# --------------------------------------------------------------------------

module "cluster" {
  source = "../../modules/osd-cluster"

  name           = var.cluster_name
  cloud_region   = var.gcp_region
  gcp_project_id = var.gcp_project_id

  openshift_version = var.openshift_version
  compute_nodes     = 3
  ccs_enabled       = true

  private = true

  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  private_service_connect = {
    service_attachment_subnet = module.osd_vpc.psc_subnet
  }

  security = {
    secure_boot = true
  }

  machine_pools = var.machine_pools
}
