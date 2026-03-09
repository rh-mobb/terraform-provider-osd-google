output "vpc_name" {
  description = "VPC network name for osdgoogle_cluster gcp_network"
  value       = google_compute_network.vpc.name
}

output "vpc_id" {
  description = "VPC network ID (self link)"
  value       = google_compute_network.vpc.id
}

output "control_plane_subnet" {
  description = "Control plane subnet name for osdgoogle_cluster gcp_network"
  value       = google_compute_subnetwork.master.name
}

output "compute_subnet" {
  description = "Compute/worker subnet name for osdgoogle_cluster gcp_network"
  value       = google_compute_subnetwork.worker.name
}

output "psc_subnet" {
  description = "PSC subnet name for osdgoogle_cluster private_service_connect (when enable_psc is true)"
  value       = var.enable_psc ? google_compute_subnetwork.psc[0].name : null
}
