// interaction 包提供事件驱动的交互层功能
//
// 该包负责事件的定义、分发和处理，是 SingerOS 的核心交互层。
// 支持多种渠道的事件接入，并通过事件总线进行分发。
package interaction

// 事件主题常量定义
const (
	// TopicGithubIssueComment GitHub Issue 评论事件主题
	TopicGithubIssueComment = "interaction.github.issue_comment"
	// TopicGithubPullRequest GitHub Pull Request 事件主题
	TopicGithubPullRequest = "interaction.github.pull_request"
)
