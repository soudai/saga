package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Client struct {
	http *http.Client
}

func NewClient(socketPath string) *Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		},
	}

	return &Client{
		http: &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) Status(ctx context.Context) (StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://saga/status", nil)
	if err != nil {
		return StatusResponse{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return StatusResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StatusResponse{}, fmt.Errorf("status request failed: %s", resp.Status)
	}

	var payload StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return StatusResponse{}, err
	}
	return payload, nil
}

func (c *Client) Cancel(ctx context.Context, taskID int64) error {
	return c.postTaskAction(ctx, taskID, "cancel")
}

func (c *Client) Retry(ctx context.Context, taskID int64) error {
	return c.postTaskAction(ctx, taskID, "retry")
}

func (c *Client) Resume(ctx context.Context, taskID int64) error {
	return c.postTaskAction(ctx, taskID, "resume")
}

func (c *Client) postTaskAction(ctx context.Context, taskID int64, action string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://saga/tasks/%d/%s", taskID, action), nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return nil
}
