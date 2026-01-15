package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

var _ resource.Resource = &MetadataResource{}
var _ resource.ResourceWithImportState = &MetadataResource{}

func NewMetadataResource() resource.Resource {
	return &MetadataResource{}
}

type MetadataResource struct {
	client *client.Client
}

type MetadataResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Path  types.String `tfsdk:"path"`
	Value types.String `tfsdk:"value"`
}

func (r *MetadataResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metadata"
}

func (r *MetadataResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages NahCloud key-value metadata with path-based hierarchy.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the metadata entry.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The path for the metadata entry (e.g., `/config/app/setting`).",
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value for the metadata entry.",
			},
		},
	}
}

func (r *MetadataResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *MetadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MetadataResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata, err := r.client.CreateMetadata(ctx, data.Path.ValueString(), data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create metadata: %s", err))
		return
	}

	data.ID = types.StringValue(metadata.ID)
	data.Path = types.StringValue(metadata.Path)
	data.Value = types.StringValue(metadata.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetadataResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata, err := r.client.GetMetadata(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read metadata: %s", err))
		return
	}

	data.Path = types.StringValue(metadata.Path)
	data.Value = types.StringValue(metadata.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MetadataResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pathVal := data.Path.ValueString()
	value := data.Value.ValueString()

	updateReq := &client.UpdateMetadataRequest{
		Path:  &pathVal,
		Value: &value,
	}

	metadata, err := r.client.UpdateMetadata(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update metadata: %s", err))
		return
	}

	data.Path = types.StringValue(metadata.Path)
	data.Value = types.StringValue(metadata.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MetadataResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMetadata(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete metadata: %s", err))
		return
	}
}

func (r *MetadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
