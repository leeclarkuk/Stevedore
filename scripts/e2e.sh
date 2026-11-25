#!/usr/bin/env bash
# Converts every example, applies it to a kind cluster and waits for the
# Deployments to become available. Reuses the cluster if it already exists so
# repeated local runs are quick; CI starts from a fresh one.
set -euo pipefail

CLUSTER="${KIND_CLUSTER:-stevedore-e2e}"
BIN="${BIN:-./bin/stevedore}"
OUT="$(mktemp -d)"
trap 'rm -rf "$OUT"' EXIT

for tool in kind kubectl; do
  command -v "$tool" >/dev/null || { echo "e2e: $tool is required" >&2; exit 1; }
done

# A private kubeconfig keeps kind from changing the user's current context.
export KUBECONFIG="$OUT/kubeconfig"
if kind get clusters | grep -qx "$CLUSTER"; then
  kind export kubeconfig --name "$CLUSTER"
else
  kind create cluster --name "$CLUSTER" --wait 120s
fi
CONTEXT="kind-$CLUSTER"

for dir in examples/*/; do
  app="$(basename "$dir")"
  ns="e2e-$app"
  echo "== $app"
  "$BIN" convert -f "$dir/compose.yaml" -o "$OUT/$app"
  kubectl --context "$CONTEXT" create namespace "$ns" --dry-run=client -o yaml \
    | kubectl --context "$CONTEXT" apply -f -
  kubectl --context "$CONTEXT" -n "$ns" apply -k "$OUT/$app"
  kubectl --context "$CONTEXT" -n "$ns" wait deployment --all --for=condition=Available --timeout=180s
done

echo "e2e: all examples applied and available on $CONTEXT"
