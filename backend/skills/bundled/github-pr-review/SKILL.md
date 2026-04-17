---
name: github-pr-review
description: Review GitHub pull requests and push events using SingerOS review conventions.
version: 0.1.0
metadata:
  singeros:
    category: github
    tags:
      - github
      - pr
      - push
      - review
    always: true
    requires_tools:
      - github.pr.get_metadata
      - github.pr.get_files
      - github.repo.compare_commits
      - github.repo.get_file
      - github.pr.publish_review
---
# GitHub Code Review

## When to Use

Use this skill for:

- GitHub pull request `opened`, `reopened`, `synchronize`, or `ready_for_review`
- GitHub `push` events when code review is needed

## Operating Mode

- Start from the event brief and raw payload. Do not assume missing details.
- Treat GitHub tools as the source of truth for repository and diff state.
- Read only enough code to justify a concrete review comment.

## Review Procedure

1. For pull requests, read PR metadata first, then read the changed file list.
2. For push events, compare the `before` and `after` revisions to get the changed commits and files.
3. Prioritize files that affect runtime behavior, auth, tool execution, data flow, safety, or public APIs.
4. Open only the most relevant files. Avoid reviewing every changed file if the signal is low.
5. Write findings only when you can point to a concrete behavioral risk, regression, or maintainability hazard.
6. If there is not enough evidence for a bug, prefer a short summary over a speculative finding.

## Publishing Rules

- Do not auto-approve.
- Use `COMMENT` when the review looks acceptable or when feedback is advisory.
- Use `REQUEST_CHANGES` only when there is a concrete issue that should block merging.
- Keep the published review concise and evidence-based.
- Inline comments are optional. Use them only when a specific file location materially improves clarity.
- For push events, use the same review standards even if there is no PR review to publish.

## Review Output Style

- Start with a short overall assessment.
- If there are findings, list only the concrete ones worth acting on.
- Mention the relevant file path when it materially helps the reviewer.
- Do not invent line numbers or code paths you did not inspect.
- If there are no blocking issues, say so directly and keep the review short.

## What Good Findings Look Like

- A clear statement of the problem.
- The concrete file or code path involved.
- Why the behavior is risky or incorrect.
- A short recommendation when it is obvious.

## Constraints

- Prefer actionable findings over broad summaries.
- Do not claim a bug unless the code path is concrete.
- If repository context is insufficient, explicitly state the missing information.
- Do not publish a harsh review when the evidence is weak.
