# Contributing

Thanks for contributing to `OpenAgentCleaner`.

## Before You Start

- Open an issue for behavior changes, larger features, or new assistant support.
- Keep the project conservative about deletion safety. When in doubt, classify a path as `confirm` or `manual`.
- Avoid broad cleanup rules without a clear source or reproducible local evidence.

## Development Workflow

1. Fork the repository and create a topic branch.
2. Make focused changes.
3. Run:

```bash
make fmt
make test
make build
```

4. Update `README.md` if user-facing behavior changes.
5. Submit a pull request with:
   - what changed
   - why it is safe
   - how you verified it

## Adding a New Assistant

When adding support for another assistant:

- document where the assistant stores state on macOS
- separate `safe`, `confirm`, and `manual` targets
- explain why each path belongs in that class
- avoid deleting user-authored content automatically
- include any env var overrides that affect storage locations

## Pull Request Expectations

- Keep changes scoped.
- Prefer explicit reasoning over large path globs.
- Do not introduce destructive behavior without a confirmation boundary.
- Include tests when practical; if not, describe the verification gap.
