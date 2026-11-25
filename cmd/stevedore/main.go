// Command stevedore converts Docker Compose files into Kubernetes manifests.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/leeclarkuk/stevedore/internal/cli"
)

func main() {
	if err := cli.NewRoot().ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
