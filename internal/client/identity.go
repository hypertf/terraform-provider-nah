package client

import "context"

type Organization struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type APIKey struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (c *Client) GetOrganization(ctx context.Context) (*Organization, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/org", nil)
	if err != nil {
		return nil, err
	}
	var org Organization
	if err := handleResponse(resp, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

func (c *Client) CreateAPIKey(ctx context.Context, name string) (*APIKey, error) {
	resp, err := c.doRequest(ctx, "POST", "/v1/api-keys", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var key APIKey
	if err := handleResponse(resp, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

// GetAPIKey returns key metadata, never the original secret token.
func (c *Client) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	resp, err := c.doRequest(ctx, "GET", Route("api-keys", id), nil)
	if err != nil {
		return nil, err
	}
	var key APIKey
	if err := handleResponse(resp, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", Route("api-keys", id), nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}
