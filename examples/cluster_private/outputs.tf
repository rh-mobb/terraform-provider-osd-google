output "cluster_id" {
  value       = module.cluster.cluster_id
  description = "OCM cluster ID"
}

output "api_url" {
  value       = module.cluster.api_url
  description = "Kubernetes API URL (internal; reachable only from within the VPC or via tunnel)"
}

output "console_url" {
  value       = module.cluster.console_url
  description = "OpenShift web console URL (internal)"
}

output "domain" {
  value       = module.cluster.domain
  description = "Cluster base domain"
}

output "bastion_name" {
  value       = google_compute_instance.bastion.name
  description = "Bastion VM name (for use with gcloud compute ssh)"
}

output "bastion_zone" {
  value       = google_compute_instance.bastion.zone
  description = "Bastion VM zone"
}

output "gcp_project_id" {
  value       = var.gcp_project_id
  description = "GCP project ID (used by Makefile ssh/tunnel targets)"
}

output "ssh_cmd" {
  value       = "gcloud compute ssh ${google_compute_instance.bastion.name} --project=${var.gcp_project_id} --zone=${google_compute_instance.bastion.zone} --tunnel-through-iap"
  description = "Command to open an interactive SSH session to the bastion via IAP"
}

output "tunnel_cmd" {
  value       = "gcloud compute ssh ${google_compute_instance.bastion.name} --project=${var.gcp_project_id} --zone=${google_compute_instance.bastion.zone} --tunnel-through-iap -- -L 6443:<api-host>:6443 -N"
  description = "Template command to forward the cluster API port via the bastion. Replace <api-host> with the hostname from api_url."
}

output "admin_username" {
  value       = module.cluster.admin_username
  description = "Cluster admin username (when create_admin = true)"
}

output "admin_password" {
  value       = module.cluster.admin_password
  description = "Cluster admin password (sensitive; when create_admin = true)"
  sensitive   = true
}
