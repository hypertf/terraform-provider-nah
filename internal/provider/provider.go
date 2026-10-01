package provider

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
				MarkdownDescription: "The NahCloud API endpoint. Defaults to `https://nahcloud.com`. Can also be set via `NAH_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Required organization API token. Set here or via `NAH_TOKEN`. Never commit tokens to configuration.",
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

	if data.Endpoint.IsUnknown() || data.Token.IsUnknown() {
		resp.Diagnostics.AddError("Unknown provider configuration", "Endpoint and token must be known before applying resources. Use an existing organization API token.")
		return
	}

	endpoint := os.Getenv("NAH_ENDPOINT")
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = client.DefaultEndpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		resp.Diagnostics.AddAttributeError(path.Root("endpoint"), "Invalid endpoint", "Use an absolute http(s) URL without credentials, query parameters, or fragment.")
		return
	}
	token := os.Getenv("NAH_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}
	if strings.TrimSpace(token) == "" {
		resp.Diagnostics.AddAttributeError(path.Root("token"), "Missing API token", "Set token or NAH_TOKEN to an existing organization's API token. Anonymous organization creation is not supported.")
		return
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
		NewAPIKeyResource,
		NewNetworkResource,
		NewSubnetResource,
		NewDiskResource,
		NewDiskAttachmentResource,
		NewPolicyResource,
		NewPolicyBindingResource,
		NewLoadBalancerResource,
		NewLoadBalancerBackendResource,
	}
}

func (p *NahProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewInstanceDataSource,
		NewMetadataDataSource,
		NewBucketDataSource,
		NewObjectDataSource,
		NewOrganizationDataSource,
		NewNetworkDataSource,
		NewSubnetDataSource,
		NewDiskDataSource,
		NewDiskAttachmentDataSource,
		NewPolicyDataSource,
		NewPolicyBindingDataSource,
		NewLoadBalancerDataSource,
		NewLoadBalancerBackendDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NahProvider{
			version: version,
		}
	}
}
