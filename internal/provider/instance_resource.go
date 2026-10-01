package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Project   types.String `tfsdk:"project"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
	SubnetID  types.String `tfsdk:"subnet_id"`
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
			"project": schema.StringAttribute{Required: true, Validators: []validator.String{slugValidator}, MarkdownDescription: "Project slug (not ID). Changing it replaces the instance.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"region":  schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("us-east-1", "us-west-1", "eu-west-1", "eu-central-1", "ap-east-1")}, MarkdownDescription: "Region: us-east-1, us-west-1, eu-west-1, eu-central-1, or ap-east-1. Immutable.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the project this instance belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{apiString{min: 1, max: 255}},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Optional opaque subnet ID. Adding, changing, or removing it replaces the instance.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Validators:          []validator.String{nameValidator},
				MarkdownDescription: "Instance name, 1–255 ASCII letters, digits, underscores, or hyphens. Unique within the project.",
			},
			"cpu": schema.Int64Attribute{
				Validators:          []validator.Int64{int64validator.Between(1, 64)},
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Number of CPUs, 1–64. Defaults to 1.",
			},
			"memory_mb": schema.Int64Attribute{
				Validators:          []validator.Int64{int64validator.Between(1, 524288)},
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(512),
				MarkdownDescription: "Memory in MB, 1–524288. Defaults to 512.",
			},
			"image": schema.StringAttribute{
				Required:            true,
				Validators:          []validator.String{apiString{min: 1, max: 255}},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Image identifier, 1–255 bytes. Changing it replaces the instance.",
			},
			"status": schema.StringAttribute{
				Validators:          []validator.String{stringvalidator.OneOf("running", "stopped")},
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
		Region:   data.Region.ValueString(),
		Name:     data.Name.ValueString(),
		CPU:      int(data.CPU.ValueInt64()),
		MemoryMB: int(data.MemoryMB.ValueInt64()),
		Image:    data.Image.ValueString(),
		Status:   data.Status.ValueString(),
	}
	if !data.SubnetID.IsNull() && !data.SubnetID.IsUnknown() {
		subnetID := data.SubnetID.ValueString()
		createReq.SubnetID = &subnetID
	}

	instance, err := r.client.CreateInstance(ctx, data.Project.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create instance: %s", err))
		return
	}

	data.ID = types.StringValue(instance.ID)
	data.ProjectID = types.StringValue(instance.ProjectID)
	data.SubnetID = types.StringPointerValue(instance.SubnetID)
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

	instance, err := r.client.GetInstance(ctx, data.Project.ValueString(), data.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read instance: %s", err))
		return
	}

	data.Region = types.StringValue(instance.Region)
	data.ProjectID = types.StringValue(instance.ProjectID)
	data.SubnetID = types.StringPointerValue(instance.SubnetID)
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

	instance, err := r.client.UpdateInstance(ctx, data.Project.ValueString(), data.ID.ValueString(), updateReq)
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

	err := r.client.DeleteInstance(ctx, data.Project.ValueString(), data.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete instance: %s", err))
		return
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importScoped(ctx, req, resp, "project", "id")
}
