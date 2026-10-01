package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

type APIKeyResource struct{ client *client.Client }
type APIKeyModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
}

func NewAPIKeyResource() resource.Resource { return &APIKeyResource{} }
func (r *APIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}
func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Manages an organization API key. Tokens are returned only on creation and stored in state. Renaming replaces the key. Imported keys have a null token; it cannot be recovered. Do not manage the provider's own authentication key with this resource.", Attributes: map[string]schema.Attribute{
		"id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Stable API key ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":  schema.StringAttribute{Required: true, MarkdownDescription: "Key display name.", Validators: []validator.String{apiString{min: 1}}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"token": schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Secret token, available only for keys created by this resource.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func (r *APIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	r.client, ok = req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected client type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
	}
}
func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.CreateAPIKey(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create API key failed", err.Error())
		return
	}
	data.ID = types.StringValue(key.ID)
	data.Token = types.StringValue(key.Token)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	key, err := r.client.GetAPIKey(ctx, data.ID.ValueString())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read API key failed", err.Error())
		return
	}
	data.Name = types.StringValue(key.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
func (r *APIKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected API key update", "API key changes require replacement. Please report this provider bug.")
}
func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data APIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAPIKey(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete API key failed", err.Error())
	}
}
func (r *APIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importScoped(ctx, req, resp, "id")
}
