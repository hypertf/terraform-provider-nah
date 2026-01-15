package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

type InstanceResource struct {
	client *client.Client
}

type InstanceResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	CPU       types.Int64  `tfsdk:"cpu"`
	MemoryMB  types.Int64  `tfsdk:"memory_mb"`
	Image     types.String `tfsdk:"image"`
	Status    types.String `tfsdk:"status"`
}

func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a NahCloud compute instance.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project this instance belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the instance.",
			},
			"cpu": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "The number of CPUs for the instance. Defaults to 1.",
			},
			"memory_mb": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(512),
				MarkdownDescription: "The amount of memory in MB for the instance. Defaults to 512.",
			},
			"image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The image to use for the instance.",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("running"),
				MarkdownDescription: "The status of the instance. Valid values: `running`, `stopped`. Defaults to `running`.",
			},
		},
	}
}

func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.CreateInstanceRequest{
		ProjectID: data.ProjectID.ValueString(),
		Name:      data.Name.ValueString(),
		CPU:       int(data.CPU.ValueInt64()),
		MemoryMB:  int(data.MemoryMB.ValueInt64()),
		Image:     data.Image.ValueString(),
		Status:    data.Status.ValueString(),
	}

	instance, err := r.client.CreateInstance(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create instance: %s", err))
		return
	}

	data.ID = types.StringValue(instance.ID)
	data.ProjectID = types.StringValue(instance.ProjectID)
	data.Name = types.StringValue(instance.Name)
	data.CPU = types.Int64Value(int64(instance.CPU))
	data.MemoryMB = types.Int64Value(int64(instance.MemoryMB))
	data.Image = types.StringValue(instance.Image)
	data.Status = types.StringValue(instance.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.client.GetInstance(ctx, data.ID.ValueString())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	cpu := int(data.CPU.ValueInt64())
	memoryMB := int(data.MemoryMB.ValueInt64())
	image := data.Image.ValueString()
	status := data.Status.ValueString()

	updateReq := &client.UpdateInstanceRequest{
		Name:     &name,
		CPU:      &cpu,
		MemoryMB: &memoryMB,
		Image:    &image,
		Status:   &status,
	}

	instance, err := r.client.UpdateInstance(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update instance: %s", err))
		return
	}

	data.Name = types.StringValue(instance.Name)
	data.CPU = types.Int64Value(int64(instance.CPU))
	data.MemoryMB = types.Int64Value(int64(instance.MemoryMB))
	data.Image = types.StringValue(instance.Image)
	data.Status = types.StringValue(instance.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteInstance(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete instance: %s", err))
		return
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
