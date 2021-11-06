package main

import (
	"context"

	"github.com/unikiosk/unikiosk/pkg/cli"
)

func main() {
	ctx := context.Background()
	// errors are being printed by CLI handelers
	run(ctx)

}

func run(ctx context.Context) error {
	return cli.RunCLI(ctx)
}
