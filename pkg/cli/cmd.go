package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/unikiosk/unikiosk/pkg/cli/set"
)

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
)

const maxImageSize = MEGABYTE * 10

// RunCLI returns user CLI
func RunCLI(ctx context.Context) error {
	cmd := &cobra.Command{
		Short: "UniKiosk CLI",
		Long:  "Client utility for UniKiosk",
		Use:   "unikiosk --help",
	}

	cmd.AddCommand(set.New())

	// This will already have global config enriched with values
	return cmd.ExecuteContext(ctx)
}
