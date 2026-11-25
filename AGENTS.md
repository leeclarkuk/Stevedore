# Agent instructions for Stevedore

These instructions apply to every coding agent working in this repo, including Claude Code and Cursor.

## What this project is

Stevedore converts Docker Compose files into GitOps-ready Kubernetes manifests with platform policy applied. Read `docs/PLAN.md` for scope and `docs/adr/` for decisions before making structural changes.

## Rules

- Stay inside the non-goals in `docs/PLAN.md`. If a change would add a new resource type, command or dependency, stop and ask.
- Output must be deterministic. No timestamps, random names or map-order iteration in rendered output.
- Unsupported Compose keys are errors with a suggested fix, never silent drops or warnings.
- Never hand-edit files in `testdata/golden/`. Regenerate with `make golden` and show the diff.
- Never commit secrets, kubeconfigs or tokens. Test fixtures use obvious fake values.
- New architectural decisions get an ADR in `docs/adr/` using `TEMPLATE.md`, before the code.
- Prefer the standard library. Add a dependency only with a reason in the PR description.

## Code conventions

- Go, current stable release, formatted with `gofmt`, linted with `golangci-lint`
- Packages live under `internal/` unless they are deliberately public
- Errors wrap with context: `fmt.Errorf("render service %q: %w", name, err)`
- Table-driven tests; every bug fix adds a test that failed before the fix
- Comments explain why, not what

## Commands

```
make build     # build ./bin/stevedore
make test      # unit tests
make golden    # regenerate golden files
make lint      # golangci-lint
make e2e       # convert examples and apply to a kind cluster
```

## Workflow

- One issue per branch, named `type/short-description`, for example `feat/network-policy`
- Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- Run `make lint test` before declaring work done, and report what was verified
- Do not push, tag, release or change GitHub settings without explicit approval

## Writing style

British English, plain language, no marketing language in docs or CLI output.
