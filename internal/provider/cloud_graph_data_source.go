package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hypertf/terraform-provider-nah/internal/client"
)

type graphDataSource struct {
	kind   string
	client *client.Client
}

func NewNetworkDataSource() datasource.DataSource { return &graphDataSource{kind: "network"} }
func NewSubnetDataSource() datasource.DataSource  { return &graphDataSource{kind: "subnet"} }
func NewDiskDataSource() datasource.DataSource    { return &graphDataSource{kind: "disk"} }
func NewDiskAttachmentDataSource() datasource.DataSource {
	return &graphDataSource{kind: "disk_attachment"}
}
func NewPolicyDataSource() datasource.DataSource { return &graphDataSource{kind: "policy"} }
func NewPolicyBindingDataSource() datasource.DataSource {
	return &graphDataSource{kind: "policy_binding"}
}
func NewLoadBalancerDataSource() datasource.DataSource {
	return &graphDataSource{kind: "load_balancer"}
}
func NewLoadBalancerBackendDataSource() datasource.DataSource {
	return &graphDataSource{kind: "load_balancer_backend"}
}

func (d *graphDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.kind
}

func (d *graphDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	a := map[string]schema.Attribute{"project": schema.StringAttribute{Required: true, MarkdownDescription: "Project slug."}, "id": schema.StringAttribute{Required: true, MarkdownDescription: "Opaque resource ID."}, "project_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Opaque owning project ID."}}
	cs := func(s string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, MarkdownDescription: s}
	}
	switch d.kind {
	case "network":
		a["name"], a["region"] = cs("Network name."), cs("Network region.")
	case "subnet":
		a["network_id"] = schema.StringAttribute{Required: true, MarkdownDescription: "Opaque parent network ID."}
		a["name"], a["cidr"] = cs("Subnet name."), cs("Private IPv4 CIDR.")
	case "disk":
		a["name"], a["region"], a["type"] = cs("Disk name."), cs("Disk region."), cs("Disk type.")
		a["size_gb"] = schema.Int64Attribute{Computed: true, MarkdownDescription: "Disk size in GiB."}
	case "disk_attachment":
		a["disk_id"] = schema.StringAttribute{Required: true, MarkdownDescription: "Opaque parent disk ID."}
		a["instance_id"], a["device"] = cs("Opaque instance ID."), cs("Device name.")
	case "policy":
		a = map[string]schema.Attribute{"id": schema.StringAttribute{Required: true}, "org_id": cs("Opaque organization ID."), "name": cs("Policy name."), "description": cs("Policy description."), "effect": cs("Evaluation effect."), "actions": schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Evaluation actions."}}
	case "policy_binding":
		a = map[string]schema.Attribute{"id": schema.StringAttribute{Required: true}, "policy_id": schema.StringAttribute{Required: true}, "org_id": cs("Opaque organization ID."), "principal_type": cs("Principal type."), "principal_id": cs("Principal ID."), "target_type": cs("Target type."), "target_id": cs("Target ID.")}
	case "load_balancer":
		a["name"], a["subnet_id"], a["region"], a["protocol"], a["algorithm"], a["health_check_path"], a["status"] = cs("Load balancer name."), cs("Opaque subnet ID."), cs("Derived region."), cs("Listener protocol."), cs("Routing algorithm."), cs("Health check path."), cs("Status.")
		a["port"] = schema.Int64Attribute{Computed: true, MarkdownDescription: "Listener port."}
	case "load_balancer_backend":
		delete(a, "project_id")
		a["load_balancer_id"] = schema.StringAttribute{Required: true}
		a["instance_id"] = cs("Opaque instance ID.")
		a["port"], a["weight"] = schema.Int64Attribute{Computed: true}, schema.Int64Attribute{Computed: true}
		a["enabled"], a["healthy"] = schema.BoolAttribute{Computed: true}, schema.BoolAttribute{Computed: true}
	}
	resp.Schema = schema.Schema{MarkdownDescription: "Reads a NahCloud " + d.kind + ".", Attributes: a}
}
func (d *graphDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	var ok bool
	d.client, ok = req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected client type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
	}
}
func (d *graphDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var err error
	switch d.kind {
	case "network":
		var m networkModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Network
		v, err = d.client.GetNetwork(ctx, m.Project.ValueString(), m.ID.ValueString())
		if err == nil {
			setNetwork(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "subnet":
		var m subnetModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Subnet
		v, err = d.client.GetSubnet(ctx, m.Project.ValueString(), m.NetworkID.ValueString(), m.ID.ValueString())
		if err == nil {
			setSubnet(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk":
		var m diskModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Disk
		v, err = d.client.GetDisk(ctx, m.Project.ValueString(), m.ID.ValueString())
		if err == nil {
			setDisk(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "disk_attachment":
		var m diskAttachmentModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.DiskAttachment
		v, err = d.client.GetDiskAttachment(ctx, m.Project.ValueString(), m.DiskID.ValueString(), m.ID.ValueString())
		if err == nil {
			setDiskAttachment(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy":
		var m policyModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Policy
		v, err = d.client.GetPolicy(ctx, m.ID.ValueString())
		if err == nil {
			setPolicy(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "policy_binding":
		var m policyBindingModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.PolicyBinding
		v, err = d.client.GetPolicyBinding(ctx, m.PolicyID.ValueString(), m.ID.ValueString())
		if err == nil {
			setPolicyBinding(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer":
		var m loadBalancerModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.LoadBalancer
		v, err = d.client.GetLoadBalancer(ctx, m.Project.ValueString(), m.ID.ValueString())
		if err == nil {
			setLoadBalancer(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	case "load_balancer_backend":
		var m backendModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var v *client.Backend
		v, err = d.client.GetBackend(ctx, m.Project.ValueString(), m.LoadBalancerID.ValueString(), m.ID.ValueString())
		if err == nil {
			setBackend(&m, v)
			resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Read failed", err.Error())
	}
}
