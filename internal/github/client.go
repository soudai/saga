package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL string
	Owner   string
	Repo    string
	HTTP    *http.Client
}

type Issue struct {
	Number    int
	State     string
	Labels    []string
	Assignees []string
	Body      string
	Comments  []string
}

func NewClient(baseURL, owner, repo string, httpClient *http.Client) Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Owner:   owner,
		Repo:    repo,
		HTTP:    httpClient,
	}
}

func (c Client) ListOpenIssues(ctx context.Context) ([]Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=open", c.BaseURL, c.Owner, c.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list issues failed: %s", resp.Status)
	}

	var payload []struct {
		Number      int       `json:"number"`
		State       string    `json:"state"`
		Body        string    `json:"body"`
		PullRequest *struct{} `json:"pull_request"`
		Labels      []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Assignees []struct {
			Login string `json:"login"`
		} `json:"assignees"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var issues []Issue
	for _, item := range payload {
		if item.PullRequest != nil {
			continue
		}

		issue := Issue{
			Number: item.Number,
			State:  item.State,
			Body:   item.Body,
		}
		for _, label := range item.Labels {
			issue.Labels = append(issue.Labels, label.Name)
		}
		for _, assignee := range item.Assignees {
			issue.Assignees = append(issue.Assignees, assignee.Login)
		}
		issues = append(issues, issue)
	}
	return issues, nil
}
