# WIF Config

Creates a Workload Identity Federation (WIF) config in OCM. This is shared infrastructure required by all OSD cluster examples.

The Makefile applies this config automatically before any example (`make example.<name>`), and destroys it after the example on `make example.<name>.destroy`.

## Variables

- **`gcp_project_id`** (required) — GCP project ID.
- **`cluster_name`** (default: `my-cluster`) — Base name. WIF display name is `cluster_name-wif`.
- **`openshift_version`** (default: `4.21.3`) — OpenShift version. WIF roles use x.y only.
- **`role_prefix`** (optional) — Prefix for custom IAM roles. Defaults to `cluster_name` with hyphens/underscores stripped.

## Standalone usage

```bash
cd terraform/wif_config
terraform init
terraform apply -var="gcp_project_id=my-project" -var="cluster_name=my-cluster"
```

Or use `terraform.tfvars`:

```hcl
gcp_project_id = "my-gcp-project"
cluster_name   = "my-wif-cluster"
```
