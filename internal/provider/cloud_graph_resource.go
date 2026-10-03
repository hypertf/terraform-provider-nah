package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

type graphResource struct {
	kind   string
	client *client.Client
}

var _ resource.ResourceWithImportState = &graphResource{}

func NewNetworkResource() resource.Resource        { return &graphResource{kind: "network"} }
func NewSubnetResource() resource.Resource         { return &graphResource{kind: "subnet"} }
func NewDiskResource() resource.Resource           { return &graphResource{kind: "disk"} }
func NewDiskAttachmentResource() resource.Resource { return &graphResource{kind: "disk_attachment"} }
func NewPolicyResource() resource.Resource         { return &graphResource{kind: "policy"} }
func NewPolicyBindingResource() resource.Resource  { return &graphResource{kind: "policy_binding"} }
func NewLoadBalancerResource() resource.Resource   { return &graphResource{kind: "load_balancer"} }
func NewLoadBalancerBackendResource() resource.Resource {
	return &graphResource{kind: "load_balancer_backend"}
}

type networkModel struct {
	Project   types.String `tfsdk:"project"`
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Region    types.String `tfsdk:"region"`
}
type subnetModel struct {
	Project   types.String `tfsdk:"project"`
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	NetworkID types.String `tfsdk:"network_id"`
	Name      types.String `tfsdk:"name"`
	CIDR      types.String `tfsdk:"cidr"`
}
type diskModel struct {
	Project   types.String `tfsdk:"project"`
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Region    types.String `tfsdk:"region"`
	Type      types.String `tfsdk:"type"`
	SizeGB    types.Int64  `tfsdk:"size_gb"`
}
type diskAttachmentModel struct {
	Project    types.String `tfsdk:"project"`
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	DiskID     types.String `tfsdk:"disk_id"`
	InstanceID types.String `tfsdk:"instance_id"`
	Device     types.String `tfsdk:"device"`
}
type policyModel struct {
	ID          types.String `tfsdk:"id"`
	OrgID       types.String `tfsdk:"org_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Effect      types.String `tfsdk:"effect"`
	Actions     types.List   `tfsdk:"actions"`
}
type policyBindingModel struct {
	ID            types.String `tfsdk:"id"`
	OrgID         types.String `tfsdk:"org_id"`
	PolicyID      types.String `tfsdk:"policy_id"`
	PrincipalType types.String `tfsdk:"principal_type"`
	PrincipalID   types.String `tfsdk:"principal_id"`
	TargetType    types.String `tfsdk:"target_type"`
	TargetID      types.String `tfsdk:"target_id"`
}
type loadBalancerModel struct {
	Project         types.String `tfsdk:"project"`
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	Name            types.String `tfsdk:"name"`
	SubnetID        types.String `tfsdk:"subnet_id"`
	Region          types.String `tfsdk:"region"`
	Protocol        types.String `tfsdk:"protocol"`
	Port            types.Int64  `tfsdk:"port"`
	Algorithm       types.String `tfsdk:"algorithm"`
	HealthCheckPath types.String `tfsdk:"health_check_path"`
	Status          types.String `tfsdk:"status"`
}
type backendModel struct {
	Project        types.String `tfsdk:"project"`
	ID             types.String `tfsdk:"id"`
	LoadBalancerID types.String `tfsdk:"load_balancer_id"`
	InstanceID     types.String `tfsdk:"instance_id"`
	Port           types.Int64  `tfsdk:"port"`
	Weight         types.Int64  `tfsdk:"weight"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Healthy        types.Bool   `tfsdk:"healthy"`
}

func (r *graphResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.kind
}

func projectAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"project":    schema.StringAttribute{Required: true, Validators: []validator.String{slugValidator}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Project slug. Changing it replaces the resource."},
		"id":         schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Opaque stable resource ID."},
		"project_id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Opaque owning project ID."},
	}
}
func requiredName() schema.StringAttribute {
	return schema.StringAttribute{Required: true, Validators: []validator.String{nameValidator}, MarkdownDescription: "Resource name."}
}
func immutableID(text string) schema.StringAttribute {
	return schema.StringAttribute{Required: true, Validators: []validator.String{apiString{min: 1, max: 255}}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: text}
}
func regionAttr() schema.StringAttribute {
	return schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("us-east-1", "us-west-1", "eu-west-1", "eu-central-1", "ap-east-1")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Region. Changing it replaces the resource."}
}

func (r *graphResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	a := projectAttrs()
	switch r.kind {
	case "network":
		a["name"], a["region"] = requiredName(), regionAttr()
	case "subnet":
		a["network_id"], a["name"] = immutableID("Opaque parent network ID."), requiredName()
		a["cidr"] = schema.StringAttribute{Required: true, Validators: []validator.String{cidrValidator{}}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Canonical private IPv4 CIDR with prefix /16 through /28. Changing it replaces the subnet."}
	case "disk":
		a["name"], a["region"] = requiredName(), regionAttr()
		a["type"] = schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("standard", "ssd")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Disk type. Changing it replaces the disk."}
		a["size_gb"] = schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 16384)}, PlanModifiers: []planmodifier.Int64{replaceOnShrink{}}, MarkdownDescription: "Disk size in GiB. Expansion updates in place; shrinking replaces the disk."}
	case "disk_attachment":
		a["disk_id"] = immutableID("Opaque parent disk ID.")
		a["instance_id"] = immutableID("Opaque instance ID.")
		a["device"] = schema.StringAttribute{Required: true, Validators: []validator.String{apiString{min: 3, max: 3, pattern: regexp.MustCompile(`^vd[b-z]$`)}}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Immutable device name matching vd[b-z]."}
	case "policy":
		a = map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Opaque policy ID."},
			"org_id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Opaque authenticated organization ID."},
			"name":        requiredName(),
			"description": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{apiString{min: 0, max: 1024}}, MarkdownDescription: "Policy description."},
			"effect":      schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("allow", "deny")}, MarkdownDescription: "Evaluation effect."},
			"actions":     schema.ListAttribute{Required: true, ElementType: types.StringType, Validators: []validator.List{listvalidator.SizeBetween(1, 32)}, MarkdownDescription: "One to 32 exact operation names or `*`. Evaluation-only; policies do not authorize CRUD."},
		}
	case "policy_binding":
		a = map[string]schema.Attribute{
			"id":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Opaque binding ID."},
			"org_id":         schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Opaque authenticated organization ID."},
			"policy_id":      immutableID("Opaque parent policy ID."),
			"principal_type": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("organization", "api_key")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Immutable principal type."},
			"principal_id":   immutableID("Opaque principal ID."),
			"target_type":    schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("organization", "project", "network", "subnet", "instance", "disk", "load_balancer", "bucket")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Immutable target type."},
			"target_id":      immutableID("Opaque target ID."),
		}
	case "load_balancer":
		a["name"] = requiredName()
		a["subnet_id"] = immutableID("Opaque subnet ID.")
		a["region"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Region derived from the subnet and network."}
		a["protocol"] = schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf("http", "tcp")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "Listener protocol. Changing it replaces the load balancer."}
		a["port"] = schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}, PlanModifiers: []planmodifier.Int64{int64RequiresReplace{}}, MarkdownDescription: "Listener port. Changing it replaces the load balancer."}
		a["algorithm"] = schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("round_robin"), Validators: []validator.String{stringvalidator.OneOf("round_robin", "least_connections")}, MarkdownDescription: "Routing algorithm."}
		a["health_check_path"] = schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Validators: []validator.String{apiString{min: 0, max: 255, pattern: regexp.MustCompile(`^$|^/`)}}, MarkdownDescription: "HTTP health path beginning with `/`; must be empty for TCP."}
		a["status"] = schema.StringAttribute{Computed: true, MarkdownDescription: "Load balancer status."}
	case "load_balancer_backend":
		delete(a, "project_id")
		a["load_balancer_id"] = immutableID("Opaque parent load balancer ID.")
		a["instance_id"] = immutableID("Opaque instance ID in the same subnet as the load balancer.")
		a["port"] = schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Between(1, 65535)}, MarkdownDescription: "Backend port."}
		a["weight"] = schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(100), Validators: []validator.Int64{int64validator.Between(1, 100)}, MarkdownDescription: "Routing weight, 1–100."}
		a["enabled"] = schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether the backend receives traffic."}
		a["healthy"] = schema.BoolAttribute{Computed: true, MarkdownDescription: "True when enabled and the instance is running."}
	}
	resp.Schema = schema.Schema{MarkdownDescription: "Manages a NahCloud " + r.kind + ".", Attributes: a}
}

// int64RequiresReplace adapts the framework's replacement behavior without a separate dependency.
type int64RequiresReplace struct{}

func (int64RequiresReplace) Description(context.Context) string {
	return "Changing this value replaces the resource."
}
func (m int64RequiresReplace) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (int64RequiresReplace) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.StateValue.IsNull() && !req.PlanValue.IsUnknown() {
		resp.RequiresReplace = !req.StateValue.Equal(req.PlanValue)
	}
}

func (r *graphResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	r.client, ok = req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected client type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
	}
}

func listStrings(v types.List) []string {
	var out []string
	for _, e := range v.Elements() {
		value, ok := e.(types.String)
		if ok {
			out = append(out, value.ValueString())
		}
	}
	return out
}
func stringList(v []string) types.List {
	values := make([]attr.Value, len(v))
	for i := range v {
		values[i] = types.StringValue(v[i])
	}
	return types.ListValueMust(types.StringType, values)
}
func setNetwork(m *networkModel, v *client.Network) {
	m.ID = types.StringValue(v.ID)
	m.ProjectID = types.StringValue(v.ProjectID)
	m.Name = types.StringValue(v.Name)
	m.Region = types.StringValue(v.Region)
}
func setSubnet(m *subnetModel, v *client.Subnet) {
	m.ID = types.StringValue(v.ID)
	m.ProjectID = types.StringValue(v.ProjectID)
	m.NetworkID = types.StringValue(v.NetworkID)
	m.Name = types.StringValue(v.Name)
	m.CIDR = types.StringValue(v.CIDR)
}
func setDisk(m *diskModel, v *client.Disk) {
	m.ID = types.StringValue(v.ID)
	m.ProjectID = types.StringValue(v.ProjectID)
	m.Name = types.StringValue(v.Name)
	m.Region = types.StringValue(v.Region)
	m.Type = types.StringValue(v.Type)
	m.SizeGB = types.Int64Value(int64(v.SizeGB))
}
func setDiskAttachment(m *diskAttachmentModel, v *client.DiskAttachment) {
	m.ID = types.StringValue(v.ID)
	m.ProjectID = types.StringValue(v.ProjectID)
	m.DiskID = types.StringValue(v.DiskID)
	m.InstanceID = types.StringValue(v.InstanceID)
	m.Device = types.StringValue(v.Device)
}
func setPolicy(m *policyModel, v *client.Policy) {
	m.ID = types.StringValue(v.ID)
	m.OrgID = types.StringValue(v.OrgID)
	m.Name = types.StringValue(v.Name)
	m.Description = types.StringValue(v.Description)
	m.Effect = types.StringValue(v.Effect)
	m.Actions = stringList(v.Actions)
}
func setPolicyBinding(m *policyBindingModel, v *client.PolicyBinding) {
	m.ID = types.StringValue(v.ID)
	m.OrgID = types.StringValue(v.OrgID)
	m.PolicyID = types.StringValue(v.PolicyID)
	m.PrincipalType = types.StringValue(v.PrincipalType)
	m.PrincipalID = types.StringValue(v.PrincipalID)
	m.TargetType = types.StringValue(v.TargetType)
	m.TargetID = types.StringValue(v.TargetID)
}
func setLoadBalancer(m *loadBalancerModel, v *client.LoadBalancer) {
	m.ID = types.StringValue(v.ID)
	m.ProjectID = types.StringValue(v.ProjectID)
	m.Name = types.StringValue(v.Name)
	m.SubnetID = types.StringValue(v.SubnetID)
	m.Region = types.StringValue(v.Region)
	m.Protocol = types.StringValue(v.Protocol)
	m.Port = types.Int64Value(int64(v.Port))
	m.Algorithm = types.StringValue(v.Algorithm)
	m.HealthCheckPath = types.StringValue(v.HealthCheckPath)
	m.Status = types.StringValue(v.Status)
}
func setBackend(m *backendModel, v *client.Backend) {
	m.ID = types.StringValue(v.ID)
	m.LoadBalancerID = types.StringValue(v.LoadBalancerID)
	m.InstanceID = types.StringValue(v.InstanceID)
	m.Port = types.Int64Value(int64(v.Port))
	m.Weight = types.Int64Value(int64(v.Weight))
	m.Enabled = types.BoolValue(v.Enabled)
	m.Healthy = types.BoolValue(v.Healthy)
}

func (r *graphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var err error
	switch r.kind {
	case "network":
		var m networkModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Network
		v, err = r.client.CreateNetwork(ctx, m.Project.ValueString(), map[string]any{"name": m.Name.ValueString(), "region": m.Region.ValueString()})
		if err == nil {
			setNetwork(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "subnet":
		var m subnetModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Subnet
		v, err = r.client.CreateSubnet(ctx, m.Project.ValueString(), m.NetworkID.ValueString(), map[string]any{"name": m.Name.ValueString(), "cidr": m.CIDR.ValueString()})
		if err == nil {
			setSubnet(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk":
		var m diskModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Disk
		v, err = r.client.CreateDisk(ctx, m.Project.ValueString(), map[string]any{"name": m.Name.ValueString(), "region": m.Region.ValueString(), "type": m.Type.ValueString(), "size_gb": m.SizeGB.ValueInt64()})
		if err == nil {
			setDisk(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk_attachment":
		var m diskAttachmentModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.DiskAttachment
		v, err = r.client.CreateDiskAttachment(ctx, m.Project.ValueString(), m.DiskID.ValueString(), map[string]any{"instance_id": m.InstanceID.ValueString(), "device": m.Device.ValueString()})
		if err == nil {
			setDiskAttachment(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy":
		var m policyModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Policy
		v, err = r.client.CreatePolicy(ctx, map[string]any{"name": m.Name.ValueString(), "description": m.Description.ValueString(), "effect": m.Effect.ValueString(), "actions": listStrings(m.Actions)})
		if err == nil {
			setPolicy(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy_binding":
		var m policyBindingModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.PolicyBinding
		v, err = r.client.CreatePolicyBinding(ctx, m.PolicyID.ValueString(), map[string]any{"principal_type": m.PrincipalType.ValueString(), "principal_id": m.PrincipalID.ValueString(), "target_type": m.TargetType.ValueString(), "target_id": m.TargetID.ValueString()})
		if err == nil {
			setPolicyBinding(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer":
		var m loadBalancerModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.LoadBalancer
		v, err = r.client.CreateLoadBalancer(ctx, m.Project.ValueString(), map[string]any{"name": m.Name.ValueString(), "subnet_id": m.SubnetID.ValueString(), "protocol": m.Protocol.ValueString(), "port": m.Port.ValueInt64(), "algorithm": m.Algorithm.ValueString(), "health_check_path": m.HealthCheckPath.ValueString()})
		if err == nil {
			setLoadBalancer(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer_backend":
		var m backendModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Backend
		v, err = r.client.CreateBackend(ctx, m.Project.ValueString(), m.LoadBalancerID.ValueString(), map[string]any{"instance_id": m.InstanceID.ValueString(), "port": m.Port.ValueInt64(), "weight": m.Weight.ValueInt64(), "enabled": m.Enabled.ValueBool()})
		if err == nil {
			setBackend(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Create failed", err.Error())
	}
}

func (r *graphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var err error
	switch r.kind {
	case "network":
		var m networkModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Network
		v, err = r.client.GetNetwork(ctx, m.Project.ValueString(), m.ID.ValueString())
		if err == nil {
			setNetwork(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "subnet":
		var m subnetModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Subnet
		v, err = r.client.GetSubnet(ctx, m.Project.ValueString(), m.NetworkID.ValueString(), m.ID.ValueString())
		if err == nil {
			setSubnet(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk":
		var m diskModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Disk
		v, err = r.client.GetDisk(ctx, m.Project.ValueString(), m.ID.ValueString())
		if err == nil {
			setDisk(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk_attachment":
		var m diskAttachmentModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.DiskAttachment
		v, err = r.client.GetDiskAttachment(ctx, m.Project.ValueString(), m.DiskID.ValueString(), m.ID.ValueString())
		if err == nil {
			setDiskAttachment(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy":
		var m policyModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Policy
		v, err = r.client.GetPolicy(ctx, m.ID.ValueString())
		if err == nil {
			setPolicy(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy_binding":
		var m policyBindingModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.PolicyBinding
		v, err = r.client.GetPolicyBinding(ctx, m.PolicyID.ValueString(), m.ID.ValueString())
		if err == nil {
			setPolicyBinding(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer":
		var m loadBalancerModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.LoadBalancer
		v, err = r.client.GetLoadBalancer(ctx, m.Project.ValueString(), m.ID.ValueString())
		if err == nil {
			setLoadBalancer(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer_backend":
		var m backendModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Backend
		v, err = r.client.GetBackend(ctx, m.Project.ValueString(), m.LoadBalancerID.ValueString(), m.ID.ValueString())
		if err == nil {
			setBackend(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	}
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read failed", err.Error())
	}
}

func (r *graphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var err error
	switch r.kind {
	case "network":
		var m networkModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Network
		v, err = r.client.UpdateNetwork(ctx, m.Project.ValueString(), m.ID.ValueString(), map[string]any{"name": m.Name.ValueString()})
		if err == nil {
			setNetwork(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "subnet":
		var m subnetModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Subnet
		v, err = r.client.UpdateSubnet(ctx, m.Project.ValueString(), m.NetworkID.ValueString(), m.ID.ValueString(), map[string]any{"name": m.Name.ValueString()})
		if err == nil {
			setSubnet(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk":
		var m diskModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Disk
		v, err = r.client.UpdateDisk(ctx, m.Project.ValueString(), m.ID.ValueString(), map[string]any{"name": m.Name.ValueString(), "size_gb": m.SizeGB.ValueInt64()})
		if err == nil {
			setDisk(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy":
		var m policyModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Policy
		v, err = r.client.UpdatePolicy(ctx, m.ID.ValueString(), map[string]any{"name": m.Name.ValueString(), "description": m.Description.ValueString(), "effect": m.Effect.ValueString(), "actions": listStrings(m.Actions)})
		if err == nil {
			setPolicy(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer":
		var m loadBalancerModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.LoadBalancer
		v, err = r.client.UpdateLoadBalancer(ctx, m.Project.ValueString(), m.ID.ValueString(), map[string]any{"name": m.Name.ValueString(), "algorithm": m.Algorithm.ValueString(), "health_check_path": m.HealthCheckPath.ValueString()})
		if err == nil {
			setLoadBalancer(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer_backend":
		var m backendModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Backend
		v, err = r.client.UpdateBackend(ctx, m.Project.ValueString(), m.LoadBalancerID.ValueString(), m.ID.ValueString(), map[string]any{"port": m.Port.ValueInt64(), "weight": m.Weight.ValueInt64(), "enabled": m.Enabled.ValueBool()})
		if err == nil {
			setBackend(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	default:
		resp.Diagnostics.AddError("Unexpected update", r.kind+" is immutable and must be replaced")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Update failed", err.Error())
	}
}

func (r *graphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var err error
	switch r.kind {
	case "network":
		var m networkModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeleteNetwork(ctx, m.Project.ValueString(), m.ID.ValueString())
		}
	case "subnet":
		var m subnetModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeleteSubnet(ctx, m.Project.ValueString(), m.NetworkID.ValueString(), m.ID.ValueString())
		}
	case "disk":
		var m diskModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeleteDisk(ctx, m.Project.ValueString(), m.ID.ValueString())
		}
	case "disk_attachment":
		var m diskAttachmentModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeleteDiskAttachment(ctx, m.Project.ValueString(), m.DiskID.ValueString(), m.ID.ValueString())
		}
	case "policy":
		var m policyModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeletePolicy(ctx, m.ID.ValueString())
		}
	case "policy_binding":
		var m policyBindingModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeletePolicyBinding(ctx, m.PolicyID.ValueString(), m.ID.ValueString())
		}
	case "load_balancer":
		var m loadBalancerModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeleteLoadBalancer(ctx, m.Project.ValueString(), m.ID.ValueString())
		}
	case "load_balancer_backend":
		var m backendModel
		resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
		if !resp.Diagnostics.HasError() {
			err = r.client.DeleteBackend(ctx, m.Project.ValueString(), m.LoadBalancerID.ValueString(), m.ID.ValueString())
		}
	}
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete failed", err.Error())
	}
}

func (r *graphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	switch r.kind {
	case "subnet":
		importScoped(ctx, req, resp, "project", "network_id", "id")
	case "disk_attachment":
		importScoped(ctx, req, resp, "project", "disk_id", "id")
	case "policy":
		importScoped(ctx, req, resp, "id")
	case "policy_binding":
		importScoped(ctx, req, resp, "policy_id", "id")
	case "load_balancer_backend":
		importScoped(ctx, req, resp, "project", "load_balancer_id", "id")
	default:
		importScoped(ctx, req, resp, "project", "id")
	}
}
