package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultEndpoint = "https://nahcloud.com"

// Client is the NahCloud API client
type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

// NewClient creates a new NahCloud API client
func NewClient(endpoint, token string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &Client{
		endpoint: endpoint,
		token:    token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Project represents a NahCloud project
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Instance represents a NahCloud compute instance
type Instance struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	CPU       int       `json:"cpu"`
	MemoryMB  int       `json:"memory_mb"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Metadata represents NahCloud key-value metadata
type Metadata struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Bucket represents a NahCloud storage bucket
type Bucket struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Object represents a NahCloud storage object
type Object struct {
	ID        string    `json:"id"`
	BucketID  string    `json:"bucket_id"`
	Path      string    `json:"path"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

func handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Project methods

func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/projects", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var project Project
	if err := handleResponse(resp, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/projects/"+id, nil)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := handleResponse(resp, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *Client) UpdateProject(ctx context.Context, id, name string) (*Project, error) {
	resp, err := c.doRequest(ctx, "PATCH", "/v1/projects/"+id, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var project Project
	if err := handleResponse(resp, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/projects/"+id, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// Instance methods

type CreateInstanceRequest struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	CPU       int    `json:"cpu"`
	MemoryMB  int    `json:"memory_mb"`
	Image     string `json:"image"`
	Status    string `json:"status,omitempty"`
}

type UpdateInstanceRequest struct {
	Name     *string `json:"name,omitempty"`
	CPU      *int    `json:"cpu,omitempty"`
	MemoryMB *int    `json:"memory_mb,omitempty"`
	Image    *string `json:"image,omitempty"`
	Status   *string `json:"status,omitempty"`
}

func (c *Client) CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*Instance, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/instances", req)
	if err != nil {
		return nil, err
	}
	var instance Instance
	if err := handleResponse(resp, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) GetInstance(ctx context.Context, id string) (*Instance, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/instances/"+id, nil)
	if err != nil {
		return nil, err
	}
	var instance Instance
	if err := handleResponse(resp, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) UpdateInstance(ctx context.Context, id string, req *UpdateInstanceRequest) (*Instance, error) {
	resp, err := c.doRequest(ctx, "PATCH", "/v1/instances/"+id, req)
	if err != nil {
		return nil, err
	}
	var instance Instance
	if err := handleResponse(resp, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) DeleteInstance(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/instances/"+id, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// Metadata methods

func (c *Client) CreateMetadata(ctx context.Context, path, value string) (*Metadata, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/metadata", map[string]string{"path": path, "value": value})
	if err != nil {
		return nil, err
	}
	var metadata Metadata
	if err := handleResponse(resp, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (c *Client) GetMetadata(ctx context.Context, id string) (*Metadata, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/metadata/"+id, nil)
	if err != nil {
		return nil, err
	}
	var metadata Metadata
	if err := handleResponse(resp, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

type UpdateMetadataRequest struct {
	Path  *string `json:"path,omitempty"`
	Value *string `json:"value,omitempty"`
}

func (c *Client) UpdateMetadata(ctx context.Context, id string, req *UpdateMetadataRequest) (*Metadata, error) {
	resp, err := c.doRequest(ctx, "PATCH", "/v1/metadata/"+id, req)
	if err != nil {
		return nil, err
	}
	var metadata Metadata
	if err := handleResponse(resp, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (c *Client) DeleteMetadata(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/metadata/"+id, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// Bucket methods

func (c *Client) CreateBucket(ctx context.Context, name string) (*Bucket, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/buckets", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var bucket Bucket
	if err := handleResponse(resp, &bucket); err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (c *Client) GetBucket(ctx context.Context, id string) (*Bucket, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/buckets/"+id, nil)
	if err != nil {
		return nil, err
	}
	var bucket Bucket
	if err := handleResponse(resp, &bucket); err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (c *Client) UpdateBucket(ctx context.Context, id, name string) (*Bucket, error) {
	resp, err := c.doRequest(ctx, "PATCH", "/v1/buckets/"+id, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var bucket Bucket
	if err := handleResponse(resp, &bucket); err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (c *Client) DeleteBucket(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/buckets/"+id, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// Object methods

type CreateObjectRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type UpdateObjectRequest struct {
	Path    *string `json:"path,omitempty"`
	Content *string `json:"content,omitempty"`
}

func (c *Client) CreateObject(ctx context.Context, bucketID string, req *CreateObjectRequest) (*Object, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/bucket/"+bucketID+"/objects", req)
	if err != nil {
		return nil, err
	}
	var object Object
	if err := handleResponse(resp, &object); err != nil {
		return nil, err
	}
	return &object, nil
}

func (c *Client) GetObject(ctx context.Context, bucketID, id string) (*Object, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/bucket/"+bucketID+"/objects/"+id, nil)
	if err != nil {
		return nil, err
	}
	var object Object
	if err := handleResponse(resp, &object); err != nil {
		return nil, err
	}
	return &object, nil
}

func (c *Client) UpdateObject(ctx context.Context, bucketID, id string, req *UpdateObjectRequest) (*Object, error) {
	resp, err := c.doRequest(ctx, "PATCH", "/v1/bucket/"+bucketID+"/objects/"+id, req)
	if err != nil {
		return nil, err
	}
	var object Object
	if err := handleResponse(resp, &object); err != nil {
		return nil, err
	}
	return &object, nil
}

func (c *Client) DeleteObject(ctx context.Context, bucketID, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/bucket/"+bucketID+"/objects/"+id, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}
