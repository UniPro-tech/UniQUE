# AI Agent Guide

Rules this repository's AI agents must follow when creating Git commits, pull requests, and issues.

## 1. Git Attribution

### Attribution Format

```text
AIL-{number}
Assisted-by: <model-name> by <tool/agent name>
```

### AIL — AI Influence Level

`AIL-{number}` is the **AI Influence Level** by Daniel Miessler
(<https://danielmiessler.com/blog/ai-influence-level-ail>). It records how much
of the work the AI did and where the idea came from. It is **not** a ticket,
issue, or branch identifier.

Scale (v2.0): read the list top-down and stop at the first line that is true of
this change.

| Level   | Name                                 | Meaning                                                      |
| ------- | ------------------------------------ | ------------------------------------------------------------ |
| `AIL-0` | Human Created, No AI Involved        | Nothing AI-based touched the work.                           |
| `AIL-1` | Human Created, Minor AI Assistance   | AI only tidied: proofreading, grammar, cleanup.              |
| `AIL-2` | Human Created, Major AI Augmentation | AI added to work a human wrote; the human still authored it. |
| `AIL-3` | AI Created, Human Full Structure     | AI wrote it inside a structure a human fully laid out.       |
| `AIL-4` | AI Created, Human Basic Idea         | AI wrote it from a basic idea a human brought.               |
| `AIL-5` | AI Created, Human Request            | A human only asked; the AI supplied the idea itself.         |
| `AIL-6` | AI Initiated, No Human Involvement   | Nobody asked; the AI decided the work should exist.          |

Assignment rules:

- Self-assess honestly for each commit and PR from the actual process. Never
  inflate or deflate the level.
- Typical agent work: the user gives the requirement and the agent designs and
  writes it → `AIL-4`. The user specifies the full structure and the agent
  fills it in → `AIL-3`. The user says "do something about X" and the agent
  invents the approach → `AIL-5`. Use `AIL-6` only for work no human requested.
- Formatting, frontmatter, art, and publishing do not move the level. For a
  mixed change, rate the main substance of that commit/PR.
- Do not infer the level from a branch name, issue number, or commit history.
- If the level genuinely cannot be assessed, omit the AIL line and keep
  `Assisted-by:`.

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
Signed-off-by: Yuito Akatsuki <yuito@yuito-it.jp>
```

Before creating a commit, read the final commit message back and confirm these
rules are satisfied.

### GPG Signing and Sign-off (Mandatory)

Every commit must be **both GPG-signed and signed off**. This is not optional,
even when the repository does not enforce it.

- GPG sign every commit: `git commit -S`, or rely on `commit.gpgsign=true`.
  Verify with `git log --show-signature -1`.
- Add a sign-off: `git commit -s`, which appends
  `Signed-off-by: <name> <email>`. The sign-off identity must match the
  commit author/committer identity (DCO).
- Never pass `--no-gpg-sign`, never create an unsigned commit, and never strip
  an existing `Signed-off-by` trailer.
- If signing fails (no key, wrong key, passphrase prompt, expired key, or
  `gpg failed to sign the data`), **stop and report the failure to the user**.
  Never fall back to an unsigned or unsigned-off commit to "get past" it.
- On `--amend`, `rebase`, `cherry-pick`, or squash, keep both: amend with
  `-S -s`, rebase with `--gpg-sign`, and retain the `Signed-off-by` trailer when
  rewriting messages.
- Trailer order at the end of the message body: `AIL-{number}`,
  `Assisted-by: Particle AI by Claude Code`, then `Signed-off-by: ...`.
- Before finishing, verify: `git log -1 --format='%G?%n%B'` shows `G` (valid
  signature) and the body contains `Signed-off-by:`.

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
behavior: construct and verify the final text explicitly. Every commit must be
GPG-signed and carry a `Signed-off-by` trailer; a commit that fails either check
must not be pushed.

If the AIL identifier cannot be reliably inferred, still include:

```text
Assisted-by: Particle AI by Claude Code
```
