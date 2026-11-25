# ADR 0002: Output layout and labels

Status: proposed
Date: 2026-11-25

## Context

`stevedore convert` writes a Kustomize base that teams commit to a GitOps repository. Once teams have committed that output, changing a file name or a label is a breaking change: Argo CD and Flux see deleted and recreated objects, and a changed Deployment selector cannot be applied at all, because selectors are immutable.

The Phase 1 scaffold already fixes a layout and a label set in `testdata/golden/`. This ADR records them so that later resource types follow the same rules and nobody changes them by accident.

Constraints from `docs/PLAN.md`:

- output must be deterministic, so diffs between runs are reviewable
- every generated field must be traceable to a Compose key, a policy rule or a default
- environments become Kustomize overlays, so the base must be environment-neutral

## Decision

### Layout

- The output directory is a flat Kustomize base with no subdirectories.
- Each object is in its own file named `<service>-<kind>.yaml`, where `<kind>` is the Kubernetes kind in lower case, for example `web-deployment.yaml`, `web-service.yaml` and `web-configmap.yaml`.
- Objects that belong to the project rather than to one service, such as a NetworkPolicy from a Compose network, use `<compose-name>-<kind>.yaml`.
- `kustomization.yaml` lists every file under `resources`, sorted by file name.
- The base sets no namespace. Overlays set it.

### Object names

- An object is named after its Compose service or Compose key, unchanged. There are no prefixes, suffixes or hashes.
- Objects of different kinds may share a name, such as a Deployment and a Service both called `web`.

### Labels

Every object, and every pod template, gets these labels:

| Label | Value |
|---|---|
| `app.kubernetes.io/name` | Compose service name |
| `app.kubernetes.io/instance` | Compose project name |
| `app.kubernetes.io/managed-by` | `stevedore` |

Selectors, both Deployment `matchLabels` and Service `selector`, use only `app.kubernetes.io/name` and `app.kubernetes.io/instance`. These two identify a workload and never change for it. Selectors leave out `managed-by` and anything else that a stowage plan or overlay might want to change later.

### Rejected

- **`app.kubernetes.io/version` from the image tag.** Every image bump would then change a label on every object, which adds noise to diffs. The tag is already visible in the container spec.
- **One file per service holding several objects.** A shorter file listing, but diffs and `stevedore explain` are easier to follow with one object per file.
- **Hashed or prefixed names.** These are not readable and not traceable to Compose, and Kustomize overlays can add a `namePrefix` where a team wants one.

## Consequences

- Golden-file diffs show exactly which object changed, by file name.
- Phase 2 stowage plans can add labels freely, but must not touch selectors. A test should enforce this when policy lands.
- Renaming a Compose service renames its objects. GitOps tools will delete and recreate them, which is the expected behaviour but worth documenting for users.
- The `instance` label comes from the Compose project name. When `name:` is absent, compose-go derives the name from the directory, so the same file converted from a different directory gives different output. We should either require `name:` or document the risk; this is still to decide, before Phase 1 is done.
- Changing anything in this ADR after the first release is a breaking change and needs a new ADR and a release note.
