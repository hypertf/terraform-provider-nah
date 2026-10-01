package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

type OrganizationDataSource struct{ client *client.Client }
type OrganizationModel struct {
	ID   types.String `tfsdk:"id"`
	Slug types.String `tfsdk:"slug"`
	Name types.String `tfsdk:"name"`
}

func NewOrganizationDataSource() datasource.DataSource { return &OrganizationDataSource{} }
func (d *OrganizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}
func (d *OrganizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "The organization selected by the provider's API token.", Attributes: map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true, MarkdownDescription: "Organization ID."},
		"slug": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization slug."},
		"name": schema.StringAttribute{Computed: true, MarkdownDescription: "Organization display name."},
	}}
}
func (d *OrganizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected client type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
	}
}
func (d *OrganizationDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	org, err := d.client.GetOrganization(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read organization failed", err.Error())
		return
	}
	data := OrganizationModel{types.StringValue(org.ID), types.StringValue(org.Slug), types.StringValue(org.Name)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
