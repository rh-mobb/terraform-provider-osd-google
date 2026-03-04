# OSD Google Provider Examples

These examples use the **local development build** of the provider. Before running any example:

## 1. Install the provider locally

From the repository root:

```bash
make install
```

This builds the provider and installs it to `~/.terraform.d/plugins/terraform.local/local/osd-google/`. Terraform will discover it automatically when using the `terraform.local/local/osd-google` source.

## 2. Set up authentication

Export your OCM credentials:

```bash
# Option A: Token (recommended)
export OSDGOOGLE_TOKEN="your-token-from-console.redhat.com"

# Option B: Client credentials
export OSDGOOGLE_CLIENT_ID="your-client-id"
export OSDGOOGLE_CLIENT_SECRET="your-client-secret"
```

## 3. Run an example

```bash
cd cluster_basic
terraform init
terraform plan
terraform apply
```

## Production use

When the provider is published to the Terraform Registry, change the `source` in each example from:

```hcl
source = "terraform.local/local/osd-google"
```

to:

```hcl
source = "registry.terraform.io/redhat/osd-google"
```

Then run `terraform init -upgrade` to download the published provider.
