package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultEndpoint = "https://nahcloud.com"

// Client is the NahCloud API client.
type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

// NewClient creates a new NahCloud API client.
func NewClient(endpoint, token string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		httpClient: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

// Project represents a NahCloud project.
type Project struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Instance represents a NahCloud compute instance.
type Instance struct {
	ID        string    `json:"id"`
	Region    string    `json:"region"`
	ProjectID string    `json:"project_id"`
	SubnetID  *string   `json:"subnet_id"`
	Name      string    `json:"name"`
	CPU       int       `json:"cpu"`
	MemoryMB  int       `json:"memory_mb"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Metadata represents NahCloud key-value metadata.
type Metadata struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Bucket represents a NahCloud storage bucket.
type Bucket struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Object represents a NahCloud storage object.
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

type APIError struct {
	StatusCode int
	Method     string
	Path       string
}

func (e *APIError) Error() string {
	hint := ""
	switch e.StatusCode {
	case 400:
		hint = " Check field constraints; projects must be empty before deletion."
	case 401, 403:
		hint = " Check NAH_TOKEN and organization access."
	case 404:
		hint = " Check the project slug and parent/resource identifiers."
	case 409:
		hint = " A resource with that unique name or path already exists; import it or choose a different value."
	case 429:
		hint = " API rate limit reached; retry after the server's rate-limit window."
	}
	return fmt.Sprintf("NahCloud %s %s: HTTP %d (%s).%s", e.Method, e.Path, e.StatusCode, http.StatusText(e.StatusCode), hint)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// Route escapes each identifier independently; logical object paths belong in JSON.
func Route(parts ...string) string {
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return "/v1/" + strings.Join(parts, "/")
}

func handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	// Never include response bodies in diagnostics: they can echo secrets or content.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return &APIError{resp.StatusCode, resp.Request.Method, resp.Request.URL.EscapedPath()}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20+1))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) > 32<<20 {
		return fmt.Errorf("NahCloud response exceeds 32 MiB")
	}
	if result != nil {
		if len(body) == 0 {
			return fmt.Errorf("NahCloud returned an empty resource response")
		}
		if strings.TrimSpace(string(body)) == "null" {
			return fmt.Errorf("NahCloud returned a null resource response")
		}
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Project methods

func (c *Client) CreateProject(ctx context.Context, slug, name string) (*Project, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/projects", map[string]string{"slug": slug, "name": name})
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
	resp, err := c.doRequest(ctx, "GET", Route("projects", id), nil)
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
	resp, err := c.doRequest(ctx, "PATCH", Route("projects", id), map[string]string{"name": name})
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
	resp, err := c.doRequest(ctx, "DELETE", Route("projects", id), nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// Instance methods

type CreateInstanceRequest struct {
	Region   string  `json:"region"`
	SubnetID *string `json:"subnet_id,omitempty"`
	Name     string  `json:"name"`
	CPU      int     `json:"cpu"`
	MemoryMB int     `json:"memory_mb"`
	Image    string  `json:"image"`
	Status   string  `json:"status,omitempty"`
}

type UpdateInstanceRequest struct {
	Name     *string `json:"name,omitempty"`
	CPU      *int    `json:"cpu,omitempty"`
	MemoryMB *int    `json:"memory_mb,omitempty"`
	Image    *string `json:"image,omitempty"`
	Status   *string `json:"status,omitempty"`
}

func (c *Client) CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) (*Instance, error) {
	resp, err := c.doRequest(ctx, "POST", Route("projects", project, "instances"), req)
	if err != nil {
		return nil, err
	}
	var instance Instance
	if err := handleResponse(resp, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) GetInstance(ctx context.Context, project, id string) (*Instance, error) {
	resp, err := c.doRequest(ctx, "GET", Route("projects", project, "instances", id), nil)
	if err != nil {
		return nil, err
	}
	var instance Instance
	if err := handleResponse(resp, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) UpdateInstance(ctx context.Context, project, id string, req *UpdateInstanceRequest) (*Instance, error) {
	resp, err := c.doRequest(ctx, "PATCH", Route("projects", project, "instances", id), req)
	if err != nil {
		return nil, err
	}
	var instance Instance
	if err := handleResponse(resp, &instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func (c *Client) DeleteInstance(ctx context.Context, project, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", Route("projects", project, "instances", id), nil)
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
	resp, err := c.doRequest(ctx, "GET", Route("metadata", id), nil)
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
	resp, err := c.doRequest(ctx, "PATCH", Route("metadata", id), req)
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
	resp, err := c.doRequest(ctx, "DELETE", Route("metadata", id), nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// Bucket methods

func (c *Client) CreateBucket(ctx context.Context, project, name string) (*Bucket, error) {
	resp, err := c.doRequest(ctx, "POST", Route("projects", project, "buckets"), map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var bucket Bucket
	if err := handleResponse(resp, &bucket); err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (c *Client) GetBucket(ctx context.Context, project, id string) (*Bucket, error) {
	resp, err := c.doRequest(ctx, "GET", Route("projects", project, "buckets-by-id", id), nil)
	if err != nil {
		return nil, err
	}
	var bucket Bucket
	if err := handleResponse(resp, &bucket); err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (c *Client) UpdateBucket(ctx context.Context, project, id, name string) (*Bucket, error) {
	resp, err := c.doRequest(ctx, "PATCH", Route("projects", project, "buckets-by-id", id), map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var bucket Bucket
	if err := handleResponse(resp, &bucket); err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (c *Client) DeleteBucket(ctx context.Context, project, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", Route("projects", project, "buckets-by-id", id), nil)
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

func (c *Client) CreateObject(ctx context.Context, project, bucketID string, req *CreateObjectRequest) (*Object, error) {
	resp, err := c.doRequest(ctx, "POST", Route("projects", project, "buckets-by-id", bucketID, "objects"), req)
	if err != nil {
		return nil, err
	}
	var object Object
	if err := handleResponse(resp, &object); err != nil {
		return nil, err
	}
	return &object, nil
}

func (c *Client) GetObject(ctx context.Context, project, bucketID, id string) (*Object, error) {
	resp, err := c.doRequest(ctx, "GET", Route("projects", project, "buckets-by-id", bucketID, "objects", id), nil)
	if err != nil {
		return nil, err
	}
	var object Object
	if err := handleResponse(resp, &object); err != nil {
		return nil, err
	}
	return &object, nil
}

func (c *Client) UpdateObject(ctx context.Context, project, bucketID, id string, req *UpdateObjectRequest) (*Object, error) {
	resp, err := c.doRequest(ctx, "PATCH", Route("projects", project, "buckets-by-id", bucketID, "objects", id), req)
	if err != nil {
		return nil, err
	}
	var object Object
	if err := handleResponse(resp, &object); err != nil {
		return nil, err
	}
	return &object, nil
}

func (c *Client) DeleteObject(ctx context.Context, project, bucketID, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", Route("projects", project, "buckets-by-id", bucketID, "objects", id), nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}
