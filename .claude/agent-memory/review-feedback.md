# Review Feedback

## PR #596 - Information disclosure fix (2026-08-04)

- When fixing a vulnerability class, apply the fix to ALL instances of that pattern, not just the one reported. The initial commit fixed Authorization header leaking but missed the identical User-Agent leaking pattern at the same call sites.
- PR descriptions must match the actual diff. Claiming a fix exists when the code was never changed erodes reviewer trust and can mask unresolved vulnerabilities.
- Do not add unrelated files (.gitignore changes, tooling configs) to security fix PRs. Reviewer flagged bundled .gitignore additions as out of scope.
