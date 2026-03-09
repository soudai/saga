package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/soudai/saga/internal/store"
)

type Client struct {
	http *http.Client
}

type TaskResponse = store.Task

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
		return StatusResponse{}, readErrorResponse(resp, "status request failed")
	}

	var payload StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return StatusResponse{}, err
	}
	return payload, nil
}

func (c *Client) Enqueue(ctx context.Context, repository string, issueNumber int64) (TaskResponse, error) {
	body, err := json.Marshal(EnqueueRequest{
		Repository:  repository,
		IssueNumber: issueNumber,
	})
	if err != nil {
		return TaskResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://saga/tasks", bytes.NewReader(body))
	if err != nil {
		return TaskResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return TaskResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return TaskResponse{}, readErrorResponse(resp, "enqueue request failed")
	}

	var task TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return TaskResponse{}, err
	}
	return task, nil
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
		return readErrorResponse(resp, "request failed")
	}
	return nil
}

func readErrorResponse(resp *http.Response, prefix string) error {
	var payload struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil && payload.Error != "" {
		return fmt.Errorf("%s: %s", prefix, payload.Error)
	}

	return fmt.Errorf("%s: %s", prefix, resp.Status)
}
