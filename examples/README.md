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

```bash
make build              # rebuild after any code changes
cd examples/cluster_basic
terraform plan -var="gcp_project_id=YOUR_PROJECT"
terraform apply -var="gcp_project_id=YOUR_PROJECT"
```

**cluster_wif only:** The WIF example requires a two-phase apply because the GCP module's `for_each` depends on OCM's blueprint (known only after the WIF config is created). Use the Makefile target:

```bash
make apply-wif-cluster
```

Or manually:

```bash
cd examples/cluster_wif
terraform apply -target=osdgoogle_wif_config.wif   # Phase 1
terraform apply                                     # Phase 2
```

See [cluster_wif/README.md](cluster_wif/README.md) for full documentation.

### 4. Iterate

After making code changes, rebuild and re-run:

```bash
cd ../..                # back to repo root
make build
cd examples/cluster_wif
terraform plan
```

## Available examples

| Example | Description |
|---------|-------------|
| [cluster_basic](cluster_basic) | Basic CCS cluster (uses existing osd-ccs-admin SA) |
| [cluster_admin](cluster_admin) | Cluster with HTPasswd admin user |
| [cluster_wif](cluster_wif) | CCS cluster with Workload Identity Federation (two-phase apply: `make apply-wif-cluster`) |
| [cluster_with_vpc](cluster_with_vpc) | Cluster with module-managed VPC (BYOVPC) |
| [cluster_psc](cluster_psc) | Cluster with Private Service Connect and Secure Boot |
| [cluster_shared_vpc](cluster_shared_vpc) | Cluster using a Shared VPC |
