package cli

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/unikiosk/unikiosk/pkg/grpc/models"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
	apimodels "github.com/unikiosk/unikiosk/pkg/models"
	fileutil "github.com/unikiosk/unikiosk/pkg/util/file"
)

type config struct {
	url               string
	unikioskServerUrl string
	file              string
}

// RunCLI returns user CLI
func RunCLI(ctx context.Context) error {
	var c config
	cmd := &cobra.Command{
		Short: "UniKiosk CLI",
		Long:  "Client utility for UniKiosk",
		Use:   "unikiosk --help",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, args)
		},
	}

	cmd.Flags().StringVarP(&c.url, "url", "u", "https://synpse.net", "Set desired URL to be opened")
	cmd.Flags().StringVarP(&c.unikioskServerUrl, "server", "s", "localhost:7000", "Set desired URL to be opened")
	cmd.Flags().StringVarP(&c.file, "file", "f", "", "File to serve")

	// This will already have global config enriched with values
	return cmd.ExecuteContext(ctx)
}

func run(ctx context.Context, c config, args []string) error {
	conn, err := grpc.Dial(c.unikioskServerUrl, grpc.WithInsecure())
	if err != nil {
		fmt.Println("failed to connect to unikiosk")
		return nil
	}

	client := service.NewKioskServiceClient(conn)

	payload := c.url
	if c.file != "" {
		exists, _ := fileutil.Exist(c.file)
		if !exists {
			return fmt.Errorf("file [%s] not found", c.file)
		}
		var err error
		data, err := ioutil.ReadFile(c.file)
		if err != nil {
			return err
		}

		payload = apimodels.StaticFilePrefix + `,
		` + string(data) + `
		`
	}

	kioskState := models.KioskState{
		Content: payload,
	}

	if responseMessage, e := client.StartOrUpdate(ctx, &kioskState); e != nil {
		fmt.Println("failed to send command to unikiosk")
		return nil
	} else {
		fmt.Println("Update sent...")
		fmt.Println(responseMessage)
		fmt.Println("==============")
	}

	return nil
}
