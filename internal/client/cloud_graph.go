package client

import (
	"context"
	"net/http"
	"time"
)

type Network struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Subnet struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	NetworkID string    `json:"network_id"`
	Name      string    `json:"name"`
	CIDR      string    `json:"cidr"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Disk struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Region    string    `json:"region"`
	SizeGB    int       `json:"size_gb"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DiskAttachment struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	DiskID     string    `json:"disk_id"`
	InstanceID string    `json:"instance_id"`
	Device     string    `json:"device"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Policy struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Effect      string    `json:"effect"`
	Actions     []string  `json:"actions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PolicyBinding struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	PolicyID      string    `json:"policy_id"`
	PrincipalType string    `json:"principal_type"`
	PrincipalID   string    `json:"principal_id"`
	TargetType    string    `json:"target_type"`
	TargetID      string    `json:"target_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LoadBalancer struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	SubnetID        string    `json:"subnet_id"`
	Name            string    `json:"name"`
	Region          string    `json:"region"`
	Protocol        string    `json:"protocol"`
	Port            int       `json:"port"`
	Algorithm       string    `json:"algorithm"`
	HealthCheckPath string    `json:"health_check_path"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Backend struct {
	ID             string    `json:"id"`
	LoadBalancerID string    `json:"load_balancer_id"`
	InstanceID     string    `json:"instance_id"`
	Port           int       `json:"port"`
	Weight         int       `json:"weight"`
	Enabled        bool      `json:"enabled"`
	Healthy        bool      `json:"healthy"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (c *Client) graphCreate(ctx context.Context, path string, body, result any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

func (c *Client) graphGet(ctx context.Context, path string, result any) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

func (c *Client) graphUpdate(ctx context.Context, path string, body, result any) error {
	resp, err := c.doRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

func (c *Client) graphDelete(ctx context.Context, path string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

func (c *Client) CreateNetwork(ctx context.Context, project string, body any) (*Network, error) {
	var v Network
	err := c.graphCreate(ctx, Route("projects", project, "networks"), body, &v)
	return &v, err
}
func (c *Client) GetNetwork(ctx context.Context, project, id string) (*Network, error) {
	var v Network
	err := c.graphGet(ctx, Route("projects", project, "networks", id), &v)
	return &v, err
}
func (c *Client) UpdateNetwork(ctx context.Context, project, id string, body any) (*Network, error) {
	var v Network
	err := c.graphUpdate(ctx, Route("projects", project, "networks", id), body, &v)
	return &v, err
}
func (c *Client) DeleteNetwork(ctx context.Context, project, id string) error {
	return c.graphDelete(ctx, Route("projects", project, "networks", id))
}

func (c *Client) CreateSubnet(ctx context.Context, project, networkID string, body any) (*Subnet, error) {
	var v Subnet
	err := c.graphCreate(ctx, Route("projects", project, "networks", networkID, "subnets"), body, &v)
	return &v, err
}
func (c *Client) GetSubnet(ctx context.Context, project, networkID, id string) (*Subnet, error) {
	var v Subnet
	err := c.graphGet(ctx, Route("projects", project, "networks", networkID, "subnets", id), &v)
	return &v, err
}
func (c *Client) UpdateSubnet(ctx context.Context, project, networkID, id string, body any) (*Subnet, error) {
	var v Subnet
	err := c.graphUpdate(ctx, Route("projects", project, "networks", networkID, "subnets", id), body, &v)
	return &v, err
}
func (c *Client) DeleteSubnet(ctx context.Context, project, networkID, id string) error {
	return c.graphDelete(ctx, Route("projects", project, "networks", networkID, "subnets", id))
}

func (c *Client) CreateDisk(ctx context.Context, project string, body any) (*Disk, error) {
	var v Disk
	err := c.graphCreate(ctx, Route("projects", project, "disks"), body, &v)
	return &v, err
}
func (c *Client) GetDisk(ctx context.Context, project, id string) (*Disk, error) {
	var v Disk
	err := c.graphGet(ctx, Route("projects", project, "disks", id), &v)
	return &v, err
}
func (c *Client) UpdateDisk(ctx context.Context, project, id string, body any) (*Disk, error) {
	var v Disk
	err := c.graphUpdate(ctx, Route("projects", project, "disks", id), body, &v)
	return &v, err
}
func (c *Client) DeleteDisk(ctx context.Context, project, id string) error {
	return c.graphDelete(ctx, Route("projects", project, "disks", id))
}

func (c *Client) CreateDiskAttachment(ctx context.Context, project, diskID string, body any) (*DiskAttachment, error) {
	var v DiskAttachment
	err := c.graphCreate(ctx, Route("projects", project, "disks", diskID, "attachments"), body, &v)
	return &v, err
}
func (c *Client) GetDiskAttachment(ctx context.Context, project, diskID, id string) (*DiskAttachment, error) {
	var v DiskAttachment
	err := c.graphGet(ctx, Route("projects", project, "disks", diskID, "attachments", id), &v)
	return &v, err
}
func (c *Client) DeleteDiskAttachment(ctx context.Context, project, diskID, id string) error {
	return c.graphDelete(ctx, Route("projects", project, "disks", diskID, "attachments", id))
}

func (c *Client) CreatePolicy(ctx context.Context, body any) (*Policy, error) {
	var v Policy
	err := c.graphCreate(ctx, Route("policies"), body, &v)
	return &v, err
}
func (c *Client) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	var v Policy
	err := c.graphGet(ctx, Route("policies", id), &v)
	return &v, err
}
func (c *Client) UpdatePolicy(ctx context.Context, id string, body any) (*Policy, error) {
	var v Policy
	err := c.graphUpdate(ctx, Route("policies", id), body, &v)
	return &v, err
}
func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	return c.graphDelete(ctx, Route("policies", id))
}

func (c *Client) CreatePolicyBinding(ctx context.Context, policyID string, body any) (*PolicyBinding, error) {
	var v PolicyBinding
	err := c.graphCreate(ctx, Route("policies", policyID, "bindings"), body, &v)
	return &v, err
}
func (c *Client) GetPolicyBinding(ctx context.Context, policyID, id string) (*PolicyBinding, error) {
	var v PolicyBinding
	err := c.graphGet(ctx, Route("policies", policyID, "bindings", id), &v)
	return &v, err
}
func (c *Client) DeletePolicyBinding(ctx context.Context, policyID, id string) error {
	return c.graphDelete(ctx, Route("policies", policyID, "bindings", id))
}

func (c *Client) CreateLoadBalancer(ctx context.Context, project string, body any) (*LoadBalancer, error) {
	var v LoadBalancer
	err := c.graphCreate(ctx, Route("projects", project, "load-balancers"), body, &v)
	return &v, err
}
func (c *Client) GetLoadBalancer(ctx context.Context, project, id string) (*LoadBalancer, error) {
	var v LoadBalancer
	err := c.graphGet(ctx, Route("projects", project, "load-balancers", id), &v)
	return &v, err
}
func (c *Client) UpdateLoadBalancer(ctx context.Context, project, id string, body any) (*LoadBalancer, error) {
	var v LoadBalancer
	err := c.graphUpdate(ctx, Route("projects", project, "load-balancers", id), body, &v)
	return &v, err
}
func (c *Client) DeleteLoadBalancer(ctx context.Context, project, id string) error {
	return c.graphDelete(ctx, Route("projects", project, "load-balancers", id))
}

func (c *Client) CreateBackend(ctx context.Context, project, loadBalancerID string, body any) (*Backend, error) {
	var v Backend
	err := c.graphCreate(ctx, Route("projects", project, "load-balancers", loadBalancerID, "backends"), body, &v)
	return &v, err
}
func (c *Client) GetBackend(ctx context.Context, project, loadBalancerID, id string) (*Backend, error) {
	var v Backend
	err := c.graphGet(ctx, Route("projects", project, "load-balancers", loadBalancerID, "backends", id), &v)
	return &v, err
}
func (c *Client) UpdateBackend(ctx context.Context, project, loadBalancerID, id string, body any) (*Backend, error) {
	var v Backend
	err := c.graphUpdate(ctx, Route("projects", project, "load-balancers", loadBalancerID, "backends", id), body, &v)
	return &v, err
}
func (c *Client) DeleteBackend(ctx context.Context, project, loadBalancerID, id string) error {
	return c.graphDelete(ctx, Route("projects", project, "load-balancers", loadBalancerID, "backends", id))
}
