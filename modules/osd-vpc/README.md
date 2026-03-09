# OSD VPC Module

Creates a GCP VPC and subnets for OpenShift Dedicated (OSD) clusters using BYOVPC (Bring Your Own VPC). Based on the terraform-google-osd reference in this repository.

## Resources

- **VPC** – Custom-mode network, no auto-created subnets
- **Subnets** – Control plane (master) and compute (worker)
- **Cloud Router** – For NAT
- **Cloud NAT** – Separate NAT for master and worker subnets
- **PSC** (optional) – PSC subnet, global address, forwarding rule, DNS zones for googleapis.com and gcr.io
- **Firewall** (optional) – Rules for private clusters and bastion access

## Usage

```hcl
module "osd_vpc" {
  source = "../../modules/osd-vpc"

  project_id   = var.gcp_project_id
  region       = "us-central1"
  cluster_name = "my-cluster"

  # Optional: enable PSC for private cluster
  enable_psc             = true
  enable_private_cluster = true
  psc_cidr               = "10.0.64.0/29"
}
```

## Outputs for osdgoogle_cluster

Use the outputs with the `gcp_network` and `private_service_connect` blocks:

```hcl
resource "osdgoogle_cluster" "cluster" {
  # ...
  gcp_network = {
    vpc_name             = module.osd_vpc.vpc_name
    vpc_project_id       = var.gcp_project_id
    control_plane_subnet = module.osd_vpc.control_plane_subnet
    compute_subnet       = module.osd_vpc.compute_subnet
  }

  private_service_connect = var.enable_psc ? {
    service_attachment_subnet = module.osd_vpc.psc_subnet
  } : null
}
```

## CIDR Planning

- **Master**: Default `10.0.0.0/19` (8,192 addresses)
- **Worker**: Default `10.0.32.0/19` (8,192 addresses)
- **PSC**: Default `10.0.64.0/29` – must be /29 or larger and within machine CIDR range (master + worker combined)

Ensure ranges do not overlap with existing networks.
