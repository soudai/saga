package github

import "strings"

type Selector struct {
	Labels    []string
	Assignees []string
	Commands  []string
}

func MatchesIssue(issue Issue, selector Selector) bool {
	if issue.State != "open" {
		return false
	}
	if contains(issue.Labels, "saga:blocked") {
		return false
	}
	if intersects(issue.Labels, selector.Labels) {
		return true
	}
	if intersects(issue.Assignees, selector.Assignees) {
		return true
	}
	for _, command := range selector.Commands {
		if strings.Contains(issue.Body, command) {
			return true
		}
		for _, comment := range issue.Comments {
			if strings.Contains(comment, command) {
				return true
			}
		}
	}
	return false
}

func intersects(left, right []string) bool {
	for _, item := range right {
		if contains(left, item) {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
