# Git Attribution Rules

When creating Git commits or GitHub pull requests, always apply the following attribution rules.

## Attribution Format

Use this format:

```text
AIL-{number}
Assisted-by: Particle AI by Claude Code
```

## AIL Identifier Inference

Automatically infer the relevant AIL issue identifier.

Use the following priority:

1. Current Git branch name (e.g. `feature/AIL-4-login`)
2. Related GitHub issue or pull request
3. Explicit task context in the current conversation
4. Related Git commit history

Rules:

- Extract the identifier in the format `AIL-{number}`.
- Always use the most relevant identifier associated with the current task.
- Never hardcode an issue number.
- Never fabricate or guess an identifier without sufficient evidence.
- If no identifier can be reliably determined, omit the AIL identifier.
- Do not ask the user for an identifier if it can be inferred from the available context.

## Git Commits

For every commit created by the agent:

- Include the inferred AIL identifier in the commit message body.
- Add `Assisted-by: Particle AI by Claude Code` as a Git trailer.
- Never include `Co-Authored-By` trailers.
- Never include Claude Code session URLs.
- Preserve the existing commit message subject and other relevant message content.

Example:

```text
feat: implement login functionality

AIL-4
Assisted-by: Particle AI by Claude Code
```

Before creating a commit, check the final commit message to ensure these rules are satisfied.

## Pull Requests

For every pull request created or updated by the agent:

- Include the inferred AIL identifier in the PR description.
- Add `Assisted-by: Particle AI by Claude Code` to the end of the description.
- Never include `Co-Authored-By`.
- Never include Claude Code session URLs.
- Preserve the existing PR description and its structure.
- Avoid duplicate attribution lines when updating an existing PR.

Example:

```markdown
## Summary

Implement login functionality.

## Changes

- Add login API
- Implement authentication flow

AIL-4
Assisted-by: Particle AI by Claude Code
```

## Enforcement

These rules apply to all Git commits and pull requests created or modified by the agent in this repository.

Do not rely on default Claude Code attribution behavior. Explicitly construct and verify commit messages and PR descriptions before submitting them.

If the AIL identifier cannot be reliably inferred, still include:

```text
Assisted-by: Particle AI by Claude Code
```
