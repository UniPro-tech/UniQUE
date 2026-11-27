# AI Agent Guide

Rules this repository's AI agents must follow when creating Git commits, pull requests, and issues.

## 1. Git Attribution

### Attribution Format

```text
AIL-{number}
Assisted-by: Particle AI by Claude Code
```

### AIL Identifier Inference

Infer the relevant AIL identifier using this priority:

1. Current Git branch name (e.g. `feature/AIL-4-login`)
2. Related GitHub issue or pull request
3. Explicit task context in the current conversation
4. Related Git commit history

Rules:

- Extract the identifier in the format `AIL-{number}`.
- Use the most relevant identifier associated with the current task.
- Never hardcode an issue number.
- Never fabricate or guess an identifier without sufficient evidence.
- If no identifier can be reliably determined, omit the AIL line only.
- Do not ask the user for an identifier if it can be inferred from context.

### Git Commits

For every commit created by the agent:

- Include the inferred AIL identifier in the commit message body.
- Add `Assisted-by: <model-name> by <tool/agent name>` as a Git trailer.
- Never include `Co-Authored-By` trailers.
- Never include session URLs.
- Preserve the existing commit message subject and other relevant content.

Example:

```text
feat: implement login functionality

AIL-4
Assisted-by: Particle AI by Claude Code
```

Before creating a commit, read the final commit message back and confirm these
rules are satisfied.

### Pull Requests and Issues

For every pull request or issue created or updated by the agent:

- Include the inferred AIL identifier in the description.
- Add `Assisted-by: <model-name> by <tool/agent name>` at the end of the description.
- Never include `Co-Authored-By`.
- Never include Claude Code session URLs.
- Avoid duplicate attribution lines when updating an existing PR or issue.

## 2. Templates Are Mandatory

**If a pull request template or an issue template exists in the repository, the
agent must follow it.** Do not write a free-form description while a template
is present.

### Where to look

Pull request templates (first match wins):

1. `.github/pull_request_template.md`
2. `.github/PULL_REQUEST_TEMPLATE.md`
3. `.github/PULL_REQUEST_TEMPLATE/*.md` — if several exist, pick the one
   closest to the change
4. `pull_request_template.md` / `PULL_REQUEST_TEMPLATE.md` at the repository root
5. `docs/pull_request_template.md`

Issue templates (first match wins):

1. `.github/ISSUE_TEMPLATE/*.yml` or `*.yaml` (GitHub Issue Forms)
2. `.github/ISSUE_TEMPLATE/*.md` (Markdown templates)
3. `.github/ISSUE_TEMPLATE/config.yml` — respect `blank_issues_enabled` and
   `contact_links`

If neither exists, use a short structure of your own: `## Summary`,
`## Changes`, `## Testing`.

### How to follow them

- **Read the template file before writing the description.** Read it from the
  working tree, or from the remote when creating a PR from a branch that lacks
  it (e.g. `gh api repos/{owner}/{repo}/contents/.github/pull_request_template.md`).
- Keep the template's headings, field names, order, checkboxes, and sections.
  Do not delete, rename, or reorder them.
- Fill every required field. For Issue Forms, each `body` element becomes a
  Markdown heading; `validations.required: true` items must not be empty.
- Do not delete a section because it does not apply: write "N/A" or
  "該当なし" instead, unless the template itself instructs removal.
- HTML comments (`<!-- ... -->`) are authoring hints: remove the examples,
  follow the instructions.
- Keep checkboxes intact. Tick only what is true; never tick a claim you did
  not verify.
- When several templates exist, choose the type that matches the work
  (bug report / feature request / enhancement, etc.). Do not invent a new type.
- When a PR template contains no attribution area, put the attribution block at
  the very end, after the last existing section, separated by a blank line. The
  template's structure always wins over this file's formatting preference.
- After creating or updating the PR/issue, re-read the body and confirm that
  every template heading is present.

## 3. Enforcement

These rules apply to every commit, pull request, and issue created or modified
by the agent in this repository. Do not rely on default Claude Code attribution
behavior: construct and verify the final text explicitly.

If the AIL identifier cannot be reliably inferred, still include:

```text
Assisted-by: <model-name> by <tool/agent name>
```
