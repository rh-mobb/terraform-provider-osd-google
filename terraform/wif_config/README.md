# WIF Config

Creates a Workload Identity Federation (WIF) config in OCM. This config is shared infrastructure required by all OSD cluster examples. It must be applied **in a separate Terraform run** before any example — see [Why a separate apply?](#why-a-separate-apply) below.

The Makefile orchestrates this automatically: `make example.<name>` applies `terraform/wif_config/` first, then the example; `make example.<name>.destroy` destroys the example first, then the WIF config.

## Why a separate apply?

The cluster examples (`examples/cluster`, `examples/cluster_psc`, etc.) use `data.osdgoogle_wif_config` to look up the WIF config by display name. That data source can only resolve when the WIF config **already exists** in OCM. If you ran both in one config, the data source would fail at plan time because nothing has been created yet.

In addition, OCM returns a "blueprint" (workload identity pool ID, service accounts, custom roles, IAM bindings) as computed attributes **after** the WIF config is created. The GCP module (`modules/osd-wif-gcp`) uses `for_each` over that blueprint. Terraform requires `for_each` keys to be known at plan time, but the blueprint is only available after the resource exists. Splitting into two configs avoids this chicken-and-egg: `terraform/wif_config` creates the WIF config and blueprint; the examples then look it up via the data source and provision GCP + cluster.

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
