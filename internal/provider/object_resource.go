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

var _ resource.Resource = &ObjectResource{}
var _ resource.ResourceWithImportState = &ObjectResource{}

func NewObjectResource() resource.Resource {
	return &ObjectResource{}
}

type ObjectResource struct {
	client *client.Client
}

type ObjectResourceModel struct {
	ID       types.String `tfsdk:"id"`
	BucketID types.String `tfsdk:"bucket_id"`
	Path     types.String `tfsdk:"path"`
	Content  types.String `tfsdk:"content"`
}

func (r *ObjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (r *ObjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a NahCloud storage object within a bucket. Content is stored as base64-encoded string.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the bucket this object belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The path of the object within the bucket.",
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The content of the object (base64-encoded).",
			},
		},
	}
}

func (r *ObjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ObjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.CreateObjectRequest{
		Path:    data.Path.ValueString(),
		Content: data.Content.ValueString(),
	}

	object, err := r.client.CreateObject(ctx, data.BucketID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create object: %s", err))
		return
	}

	data.ID = types.StringValue(object.ID)
	data.BucketID = types.StringValue(object.BucketID)
	data.Path = types.StringValue(object.Path)
	data.Content = types.StringValue(object.Content)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ObjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	object, err := r.client.GetObject(ctx, data.BucketID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read object: %s", err))
		return
	}

	data.BucketID = types.StringValue(object.BucketID)
	data.Path = types.StringValue(object.Path)
	data.Content = types.StringValue(object.Content)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ObjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pathVal := data.Path.ValueString()
	content := data.Content.ValueString()

	updateReq := &client.UpdateObjectRequest{
		Path:    &pathVal,
		Content: &content,
	}

	object, err := r.client.UpdateObject(ctx, data.BucketID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update object: %s", err))
		return
	}

	data.Path = types.StringValue(object.Path)
	data.Content = types.StringValue(object.Content)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ObjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteObject(ctx, data.BucketID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete object: %s", err))
		return
	}
}

func (r *ObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
