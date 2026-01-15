package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

var _ datasource.DataSource = &MetadataDataSource{}

func NewMetadataDataSource() datasource.DataSource {
	return &MetadataDataSource{}
}

type MetadataDataSource struct {
	client *client.Client
}

type MetadataDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Path      types.String `tfsdk:"path"`
	Value     types.String `tfsdk:"value"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *MetadataDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metadata"
}

func (d *MetadataDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a NahCloud metadata entry.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the metadata entry.",
			},
			"path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The path for the metadata entry.",
			},
			"value": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The value for the metadata entry.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the metadata was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the metadata was last updated.",
			},
		},
	}
}

func (d *MetadataDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *MetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MetadataDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata, err := d.client.GetMetadata(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read metadata: %s", err))
		return
	}

	data.Path = types.StringValue(metadata.Path)
	data.Value = types.StringValue(metadata.Value)
	data.CreatedAt = types.StringValue(metadata.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.UpdatedAt = types.StringValue(metadata.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
