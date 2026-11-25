# Stevedore plan

Status: draft, 25 November 2026. Owner: Lee Clark.

## What we are building

Stevedore turns Docker Compose files into GitOps-ready Kubernetes manifests, with platform policy applied and per-environment overlays. It is a single static binary that runs anywhere, including CI, and needs no Docker Desktop.

Developers keep writing Compose. Platform teams keep control of what reaches the cluster.

## Who it is for

Platform teams that:

- run Kubernetes through GitOps (Argo CD or Flux)
- have developers who already use Compose locally
- cannot or will not mandate Docker Desktop
- need to show auditors that workloads meet a stated standard

## Why it should exist

| Tool | Gap Stevedore fills |
|---|---|
| Kompose | No policy layer, no environment model, one-shot conversion |
| Compose Bridge | Tied to Docker Desktop; policy is DIY Go templates |
| Score | Requires a new spec instead of Compose |
| Tako | Same idea, abandoned since 2021 |

The three things nobody does well: versioned policy packs, a proper environment model, and day-two commands (`explain`, `diff`, `drift`).

## Non-goals

- Running containers or replacing `docker compose up`
- Inner-loop development (Skaffold, Tilt and DevSpace do this)
- Secret storage. We reference secrets; we never hold them
- Operators, CRDs or anything that runs in the cluster
- Supporting every Compose key. Unsupported keys fail loudly, never silently

## Design principles

1. **Compose is the only input.** Configuration lives in `x-stevedore` extension fields, which the Compose spec already allows.
2. **Deterministic output.** Same input, same bytes. Stable ordering, no timestamps. Diffs must be reviewable.
3. **Fail loudly.** An unsupported or ambiguous key is an error with a fix suggestion, not a warning nobody reads.
4. **Explainable.** Every generated field can be traced to a Compose key, a policy rule or a default.
5. **Boring dependencies.** Official libraries where they exist, standard library otherwise.

## Phases

### Phase 0: validate (weeks 1 to 2)

- Interview 8 to 10 platform leads. Question: how do you get from Compose to the cluster today, and what breaks?
- Mine open issues on Kompose and Compose Bridge for repeated unmet requests
- Finish name checks: GitHub org, Go module path, Homebrew, domain, UK IPO and EUIPO class 9

**Gate:** if fewer than half describe a pain Stevedore fixes, stop or pivot. Write the result in `docs/validation.md`.

### Phase 1: convert (weeks 3 to 6)

- `stevedore convert` reads Compose through `compose-go` and writes a Kustomize base
- Resources: Deployment, Service, ConfigMap, PersistentVolumeClaim, Secret references, NetworkPolicy from Compose networks
- Environments: `compose.<env>.yaml` overrides become Kustomize overlays
- Golden-file tests for every example in `examples/`
- Release pipeline: GoReleaser, signed binaries, Homebrew tap

**Done when:** three real-world Compose apps convert, apply cleanly to a kind cluster and pass `kubeconform`.

### Phase 2: stowage plans (weeks 7 to 10)

- Policy packs, called stowage plans, versioned and shipped in the binary
- `baseline` and `restricted`, aligned to Kubernetes Pod Security Standards
- Resource defaults, probes, security contexts, labels
- `stevedore explain <resource>` shows where every field came from
- Custom stowage plans loaded from a path or OCI reference

**Done when:** `restricted` output passes Pod Security Admission in enforce mode.

### Phase 3: day two (weeks 11 to 14)

- `stevedore diff` against a GitOps repo or a live cluster
- `stevedore drift` exits non-zero when the cluster no longer matches Compose
- GitHub Action wrapping `convert` and `diff` for pull requests

### Phase 4: reach (later, only if adoption warrants it)

- Read Compose Bridge templates, so Bridge users can switch at no cost
- Helm chart output
- Docs site

## Repository layout

```
cmd/stevedore/        CLI entry point (cobra)
internal/compose/     Load and validate Compose via compose-go
internal/model/       Stevedore's intermediate model
internal/render/      Model to Kubernetes objects
internal/overlay/     Environments to Kustomize overlays
internal/policy/      Stowage plans
internal/explain/     Provenance tracking
examples/             Compose apps used as fixtures
testdata/golden/      Expected output, regenerated, never hand-edited
docs/adr/             Architecture decision records
```

## Engineering setup

- Language: Go, current stable release. See ADR 0001.
- CI: GitHub Actions running `golangci-lint`, `go test ./...`, golden tests and a kind smoke test
- Commits: Conventional Commits, so release notes can be generated
- Branches: one issue per branch, squash merge to `main`
- Licence: Apache 2.0 (already in repo). Add DCO sign-off before accepting outside contributions

## Working with Claude Code and Cursor

- `AGENTS.md` is the single source of agent instructions. Cursor reads it directly; `CLAUDE.md` imports it, so the two never drift
- Use Claude Code for multi-file work: new resource types, refactors, test suites. Start in plan mode for anything touching `internal/model`
- Use Cursor for editing, review and small fixes
- Golden tests are the safety net for agent changes. If output changes, the diff in `testdata/golden` is what you review
- Every architectural decision gets an ADR before code. Agents should read `docs/adr/` first

## Decisions still to make

| Decision | Recommendation | Needed by |
|---|---|---|
| Policy engine | Go code plus CEL for custom rules; avoid Rego unless users ask | Phase 2 |
| Secrets | Emit references only; document External Secrets Operator pattern | Phase 1 |
| Image builds | Out of scope; require `image:` per service, fail if only `build:` | Phase 1 |
| Ingress | Gateway API `HTTPRoute`, not Ingress | Phase 1 |
| Governance | Personal repo now; move to an org once name is cleared | Phase 0 |

## Success measures

- Phase 1: three external teams run it on a real app
- Six months: 500 GitHub stars, 10 issues from people we have never met, one paid engagement traced to the project
- Kill signal: no outside issues or contributions three months after first release

## Risks

| Risk | Mitigation |
|---|---|
| Docker decouples Bridge from Desktop and adds policy | Lead on stowage plans and day-two commands, which Docker is least likely to prioritise |
| Compose spec and Kubernetes APIs move | `compose-go` for parsing; pin API versions; quarterly upgrade release |
| Maintenance load on one person | Tight non-goals; refuse features that add new resource types without a user asking |
| Naming collision with OpenStack stevedore | Distinct org and domain; consider `stv` as short binary alias |
