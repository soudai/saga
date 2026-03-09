package github

import "testing"

func TestMatchesIssue(t *testing.T) {
	t.Parallel()

	selector := Selector{
		Labels:    []string{"saga:ready"},
		Assignees: []string{"saga-bot"},
		Commands:  []string{"/saga run"},
	}

	tests := []struct {
		name  string
		issue Issue
		want  bool
	}{
		{
			name: "closed issue",
			issue: Issue{
				State:  "closed",
				Labels: []string{"saga:ready"},
			},
			want: false,
		},
		{
			name: "blocked issue",
			issue: Issue{
				State:  "open",
				Labels: []string{"saga:blocked", "saga:ready"},
			},
			want: false,
		},
		{
			name: "label only",
			issue: Issue{
				State:  "open",
				Labels: []string{"saga:ready"},
			},
			want: true,
		},
		{
			name: "assignee only",
			issue: Issue{
				State:     "open",
				Assignees: []string{"saga-bot"},
			},
			want: true,
		},
		{
			name: "command only",
			issue: Issue{
				State: "open",
				Comments: []Comment{
					{Body: "/saga run", Author: "soudai"},
				},
			},
			want: true,
		},
		{
			name: "bot command ignored",
			issue: Issue{
				State: "open",
				Comments: []Comment{
					{Body: "/saga run", Author: "saga-bot", AuthorIsBot: true},
				},
			},
			want: false,
		},
		{
			name: "body command ignored",
			issue: Issue{
				State: "open",
				Body:  "/saga run",
			},
			want: false,
		},
		{
			name: "no match",
			issue: Issue{
				State: "open",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := MatchesIssue(tt.issue, selector); got != tt.want {
				t.Fatalf("MatchesIssue() = %v, want %v", got, tt.want)
			}
		})
	}
}
