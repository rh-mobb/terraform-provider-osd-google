/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/redhat/terraform-provider-osd-google/provider/datasources"
	tfpschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"

	"github.com/redhat/terraform-provider-osd-google/build"
	"github.com/redhat/terraform-provider-osd-google/logging"
	"github.com/redhat/terraform-provider-osd-google/provider/cluster"
	"github.com/redhat/terraform-provider-osd-google/provider/cluster_waiter"
	"github.com/redhat/terraform-provider-osd-google/provider/dns_domain"
	"github.com/redhat/terraform-provider-osd-google/provider/machine_pool"
	"github.com/redhat/terraform-provider-osd-google/provider/wif_config"
)

// Provider is the implementation of the OSD Google provider.
type Provider struct{}

var _ tfprovider.Provider = &Provider{}

// Config contains the configuration of the provider.
type Config struct {
	URL        types.String `tfsdk:"url"`
	TokenURL   types.String `tfsdk:"token_url"`
	Token      types.String `tfsdk:"token"`
	ClientID   types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	TrustedCAs types.String `tfsdk:"trusted_cas"`
	Insecure   types.Bool   `tfsdk:"insecure"`
}

// New creates the provider.
func New() tfprovider.Provider {
	return &Provider{}
}

func (p *Provider) Metadata(ctx context.Context, req tfprovider.MetadataRequest, resp *tfprovider.MetadataResponse) {
	resp.TypeName = "osdgoogle"
	resp.Version = build.Version
	if resp.Version == "" {
		resp.Version = "0.0.1"
	}
}

func (p *Provider) Schema(ctx context.Context, req tfprovider.SchemaRequest, resp *tfprovider.SchemaResponse) {
	resp.Schema = tfpschema.Schema{
		Attributes: map[string]tfpschema.Attribute{
			"url": tfpschema.StringAttribute{
				Description: fmt.Sprintf("URL sets the base URL of the API gateway. The default is `%s`", sdk.DefaultURL),
				Optional:    true,
			},
			"token_url": tfpschema.StringAttribute{
				Description: fmt.Sprintf("TokenURL is the URL for requesting OpenID access tokens. The default is `%s`", sdk.DefaultTokenURL),
				Optional:    true,
			},
			"token": tfpschema.StringAttribute{
				Description: "Access or refresh token generated from https://console.redhat.com/openshift/token/rosa",
				Optional:    true,
				Sensitive:   true,
			},
			"client_id": tfpschema.StringAttribute{
				Description: "OpenID client identifier for client credentials authentication (alternative to token). Use with client_secret. Env: OSDGOOGLE_CLIENT_ID.",
				Optional:    true,
			},
			"client_secret": tfpschema.StringAttribute{
				Description: "OpenID client secret for client credentials authentication. Env: OSDGOOGLE_CLIENT_SECRET.",
				Optional:    true,
				Sensitive:   true,
			},
			"trusted_cas": tfpschema.StringAttribute{
				Description: "PEM encoded certificates of authorities that will be trusted. " +
					"If not specified, the system default CAs are used.",
				Optional: true,
			},
			"insecure": tfpschema.BoolAttribute{
				Description: "When set to 'true' enables insecure communication with the server. " +
					"This disables verification of TLS certificates and host names. Not recommended for production.",
				Optional: true,
			},
		},
	}
}

func (p *Provider) getAttrValueOrConfig(attr types.String, envSuffix string) (string, bool) {
	if !attr.IsNull() {
		return attr.ValueString(), true
	}
	if value, ok := os.LookupEnv(fmt.Sprintf("OSDGOOGLE_%s", envSuffix)); ok {
		return value, true
	}
	return "", false
}

func (p *Provider) Configure(ctx context.Context, req tfprovider.ConfigureRequest, resp *tfprovider.ConfigureResponse) {
	var config Config
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logger := logging.New()

	builder := sdk.NewConnectionBuilder()
	builder.Logger(logger)
	builder.Agent(fmt.Sprintf("OCM-TF-OSD-GCP/%s-%s", build.Version, build.Commit))
	if build.Version == "" {
		builder.Agent("OCM-TF-OSD-GCP/0.0.1-dev")
	}

	if url, ok := p.getAttrValueOrConfig(config.URL, "URL"); ok {
		builder.URL(url)
	}
	if tokenURL, ok := p.getAttrValueOrConfig(config.TokenURL, "TOKEN_URL"); ok {
		builder.TokenURL(tokenURL)
	}

	token, hasToken := p.getAttrValueOrConfig(config.Token, "TOKEN")
	clientID, hasClientID := p.getAttrValueOrConfig(config.ClientID, "CLIENT_ID")
	clientSecret, hasClientSecret := p.getAttrValueOrConfig(config.ClientSecret, "CLIENT_SECRET")

	if hasToken {
		builder.Tokens(token)
	} else if hasClientID {
		if !hasClientSecret {
			resp.Diagnostics.AddError(
				"client_secret required when using client_id",
				"Provide client_secret or set OSDGOOGLE_CLIENT_SECRET when using client credentials authentication.",
			)
			return
		}
		builder.Client(clientID, clientSecret)
	} else {
		resp.Diagnostics.AddError(
			"authentication required",
			"Provide token (or OSDGOOGLE_TOKEN) or client_id+client_secret (or OSDGOOGLE_CLIENT_ID and OSDGOOGLE_CLIENT_SECRET). "+
				"Get a token at https://console.redhat.com/openshift/token/rosa",
		)
		return
	}
	if trustedCAs, ok := p.getAttrValueOrConfig(config.TrustedCAs, "TRUSTED_CAS"); ok {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(trustedCAs)) {
			resp.Diagnostics.AddError(
				"the value of 'trusted_cas' doesn't contain any certificate",
				"",
			)
			return
		}
		builder.TrustedCAs(pool)
	}
	if !config.Insecure.IsNull() {
		builder.Insecure(config.Insecure.ValueBool())
	}

	connection, err := builder.BuildContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp.DataSourceData = connection
	resp.ResourceData = connection
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		cluster.New,
		cluster_waiter.New,
		dns_domain.New,
		wif_config.New,
		machine_pool.New,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewVersions,
		datasources.NewMachineTypes,
		datasources.NewRegions,
	}
}
