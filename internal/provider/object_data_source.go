package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

var _ datasource.DataSource = &ObjectDataSource{}

func NewObjectDataSource() datasource.DataSource {
	return &ObjectDataSource{}
}

type ObjectDataSource struct {
	client *client.Client
}

type ObjectDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	BucketID  types.String `tfsdk:"bucket_id"`
	Path      types.String `tfsdk:"path"`
	Content   types.String `tfsdk:"content"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *ObjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (d *ObjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a NahCloud storage object.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the object.",
			},
			"bucket_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the bucket this object belongs to.",
			},
			"path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The path of the object within the bucket.",
			},
			"content": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The content of the object (base64-encoded).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the object was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the object was last updated.",
			},
		},
	}
}

func (d *ObjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ObjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ObjectDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	object, err := d.client.GetObject(ctx, data.BucketID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read object: %s", err))
		return
	}

	data.Path = types.StringValue(object.Path)
	data.Content = types.StringValue(object.Content)
	data.CreatedAt = types.StringValue(object.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.UpdatedAt = types.StringValue(object.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
