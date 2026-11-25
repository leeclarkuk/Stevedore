// Package cli defines the stevedore command tree.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/leeclarkuk/stevedore/internal/convert"
)

// version is set at build time with -ldflags "-X .../internal/cli.version=...".
var version = "dev"

// NewRoot returns the root command. A constructor rather than a package-level
// variable keeps flag state from leaking between tests.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "stevedore",
		Short:         "Convert Docker Compose files into Kubernetes manifests",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	root.AddCommand(newConvert())
	return root
}

func newConvert() *cobra.Command {
	var file, out string

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Write a Kustomize base from a Compose file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			files, err := convert.Convert(cmd.Context(), file)
			if err != nil {
				return err
			}
			if err := convert.Write(out, files); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Wrote %d files to %s\n", len(files), out)
			return err
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "compose.yaml", "Compose file to convert")
	cmd.Flags().StringVarP(&out, "output", "o", "out", "directory to write the Kustomize base to")
	return cmd
}
