package github

import (
	"github.com/google/go-github/v78/github"

	"github.com/insmtx/SingerOS/backend/interaction"
)

// convertEvent converts GitHub events to SingerOS interaction events.
func (c *Connector) convertEvent(eventType string, event any) *interaction.Event {
	switch eventType {
	case "issue_comment":
		return c.convertIssueComment(event.(*github.IssueCommentEvent))
	case "pull_request":
		return c.convertPullRequest(event.(*github.PullRequestEvent))
	default:
		return nil
	}
}

// convertIssueComment converts GitHub IssueCommentEvent to SingerOS Event.
func (c *Connector) convertIssueComment(event *github.IssueCommentEvent) *interaction.Event {
	return &interaction.Event{
		Channel:    c.ChannelCode(),
		EventType:  EventTypeIssueComment,
		Actor:      event.GetComment().GetUser().GetLogin(),
		Repository: event.GetRepo().GetFullName(),
		Payload: map[string]interface{}{
			"issue_number": event.GetIssue().GetNumber(),
			"comment":      event.GetComment().GetBody(),
			"comment_id":   event.GetComment().GetID(),
		},
	}
}

// convertPullRequest converts GitHub PullRequestEvent to SingerOS Event.
func (c *Connector) convertPullRequest(event *github.PullRequestEvent) *interaction.Event {
	return &interaction.Event{
		Channel:    c.ChannelCode(),
		EventType:  EventTypePullRequest,
		Actor:      event.GetSender().GetLogin(),
		Repository: event.GetRepo().GetFullName(),
		Payload: map[string]interface{}{
			"pr_number": event.GetPullRequest().GetNumber(),
			"title":     event.GetPullRequest().GetTitle(),
			"body":      event.GetPullRequest().GetBody(),
			"action":    event.GetAction(),
			"state":     event.GetPullRequest().GetState(),
			"url":       event.GetPullRequest().GetHTMLURL(),
		},
	}
}
