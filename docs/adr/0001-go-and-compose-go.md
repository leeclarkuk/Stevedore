# ADR 0001: Go and compose-go

Status: accepted
Date: 2026-11-25

## Context

Stevedore must parse Compose exactly as Docker does, including interpolation, overrides, profiles and extension fields, and ship as a single static binary.

## Decision

Write Stevedore in Go. Parse Compose with `github.com/compose-spec/compose-go`, the library Docker Compose itself uses. Build Kubernetes objects with the upstream `k8s.io/api` types and serialise with `sigs.k8s.io/yaml`.

## Consequences

- Compose behaviour matches Docker's, including edge cases we would otherwise get wrong
- Single static binary, easy to ship through Homebrew and CI
- Same language as Kompose, Compose Bridge and most of the Kubernetes ecosystem, so contributors can move between them
- We inherit compose-go's release cadence and must track it
