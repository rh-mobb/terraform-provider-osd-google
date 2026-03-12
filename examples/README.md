# OSD Google Provider Examples

These examples use the [Terraform Registry](https://registry.terraform.io/providers/rh-mobb) source (`registry.terraform.io/rh-mobb/osd-google`).
For local development, use Terraform's `dev_overrides` so the local build is used without changing any example files.

## Local development setup

### 1. Build the provider and configure dev_overrides

From the repository root:

```bash
make dev-setup
```

This builds the provider binary and prints a `~/.terraformrc` snippet.
Add the printed `provider_installation` block to `~/.terraformrc` (create the file if it doesn't exist).

With `dev_overrides` active, Terraform uses the local binary directly — no `terraform init` required.

### 2. Set up authentication

Export your OCM credentials:

```bash
# Option A: Token (recommended)
export OSDGOOGLE_TOKEN="your-token-from-console.redhat.com"

# Option B: Client credentials
export OSDGOOGLE_CLIENT_ID="your-client-id"
export OSDGOOGLE_CLIENT_SECRET="your-client-secret"
```

For examples that use the Google provider (most of them), authenticate with GCP:

```bash
gcloud auth application-default login
```

### 3. Run an example

Every example target handles the full lifecycle — WIF config and cluster are created and destroyed together:

```bash
make example.cluster              # Create WIF config + cluster
make example.cluster.destroy      # Destroy cluster + WIF config
```

The Make targets infer `gcp_project_id` from `gcloud config` and `cluster_name` from your username (override with `GCP_PROJECT_ID` and `CLUSTER_NAME`). See [cluster/README.md](cluster/README.md) for full documentation.

### 4. Iterate

After making code changes, rebuild and re-run:

```bash
make build
cd examples/cluster
terraform plan
```

## Available examples

| Example | Description |
|---------|-------------|
| [cluster](cluster) | CCS cluster with WIF and cluster admin |
| [cluster_baremetal](cluster_baremetal) | Single-AZ cluster with bare metal as default compute type |
| [cluster_with_vpc](cluster_with_vpc) | Cluster with module-managed VPC (BYOVPC) |
| [cluster_psc](cluster_psc) | Cluster with Private Service Connect and Secure Boot |
| [cluster_shared_vpc](cluster_shared_vpc) | Cluster using a Shared VPC |
| [cluster_multi_az](cluster_multi_az) | Multi-AZ cluster across multiple availability zones |

WIF config (`terraform/wif_config/`) is shared infrastructure applied automatically by the Makefile before each example.
