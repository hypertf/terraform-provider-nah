package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

var _ datasource.DataSource = &InstanceDataSource{}

func NewInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

type InstanceDataSource struct {
	client *client.Client
}

type InstanceDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	CPU       types.Int64  `tfsdk:"cpu"`
	MemoryMB  types.Int64  `tfsdk:"memory_mb"`
	Image     types.String `tfsdk:"image"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *InstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (d *InstanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a NahCloud compute instance.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the instance.",
			},
			"project_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the project this instance belongs to.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the instance.",
			},
			"cpu": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of CPUs for the instance.",
			},
			"memory_mb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The amount of memory in MB for the instance.",
			},
			"image": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The image used for the instance.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the instance.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the instance was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the instance was last updated.",
			},
		},
	}
}

func (d *InstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := d.client.GetInstance(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instance: %s", err))
		return
	}

	data.ProjectID = types.StringValue(instance.ProjectID)
	data.Name = types.StringValue(instance.Name)
	data.CPU = types.Int64Value(int64(instance.CPU))
	data.MemoryMB = types.Int64Value(int64(instance.MemoryMB))
	data.Image = types.StringValue(instance.Image)
	data.Status = types.StringValue(instance.Status)
	data.CreatedAt = types.StringValue(instance.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.UpdatedAt = types.StringValue(instance.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
