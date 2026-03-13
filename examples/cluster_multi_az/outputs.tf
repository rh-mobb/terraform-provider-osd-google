output "cluster_id" {
  value       = module.cluster.cluster_id
  description = "OCM cluster ID"
}

output "api_url" {
  value       = module.cluster.api_url
  description = "Kubernetes API URL"
}

output "console_url" {
  value       = module.cluster.console_url
  description = "OpenShift web console URL"
}
