package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

var _ provider.Provider = &NahProvider{}

// NahProvider defines the provider implementation.
type NahProvider struct {
	version string
}

// NahProviderModel describes the provider data model.
type NahProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *NahProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nah"
	resp.Version = p.version
}

func (p *NahProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The NahCloud provider allows you to manage resources in NahCloud, a fake cloud API for testing Terraform tooling.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The NahCloud API endpoint. Defaults to `http://localhost:8080`. Can also be set via `NAH_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The NahCloud API token for authentication. Can also be set via `NAH_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *NahProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NahProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use environment variables as fallback
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("NAH_ENDPOINT")
	}

	token := data.Token.ValueString()
	if token == "" {
		token = os.Getenv("NAH_TOKEN")
	}

	// Create the client
	nahClient := client.NewClient(endpoint, token)

	resp.DataSourceData = nahClient
	resp.ResourceData = nahClient
}

func (p *NahProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewInstanceResource,
		NewMetadataResource,
		NewBucketResource,
		NewObjectResource,
	}
}

func (p *NahProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewInstanceDataSource,
		NewMetadataDataSource,
		NewBucketDataSource,
		NewObjectDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NahProvider{
			version: version,
		}
	}
}
