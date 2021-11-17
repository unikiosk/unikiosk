package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/unikiosk/unikiosk/pkg/api"
	"github.com/unikiosk/unikiosk/pkg/grpc/models"
	"github.com/unikiosk/unikiosk/pkg/grpc/service"
)

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
)

const maxImageSize = MEGABYTE * 10

type config struct {
	url               string
	unikioskServerUrl string
	file              string
	action            string
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
	cmd.Flags().StringVarP(&c.file, "file", "f", "", "File to write screenshoot")
	cmd.Flags().StringVarP(&c.action, "action", "a", "update", "Screen action [start,update,poweron,poweroff]")

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

	switch strings.ToLower(c.action) {
	case api.ScreenActionStart.String(), api.ScreenActionUpdate.String():
		return startOrUpdate(ctx, client, c)
	case api.ScreenActionPowerOff.String(), api.ScreenActionPowerOn.String():
		return powerOnOrOff(ctx, client, c)
	case api.ScreenActionScreenShot.String():
		return screenshot(ctx, client, c)
	default:
		return fmt.Errorf("action %s not implemented", c.action)

	}
}

func startOrUpdate(ctx context.Context, client service.KioskServiceClient, c config) error {
	payload := c.url
	kioskState := models.KioskRequest{
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

func powerOnOrOff(ctx context.Context, client service.KioskServiceClient, c config) error {
	payload := models.KioskRequest{}
	switch strings.ToLower(c.action) {
	case api.ScreenActionPowerOn.String():
		payload.Action = models.EnumScreenAction_POWERON
	case api.ScreenActionPowerOff.String():
		payload.Action = models.EnumScreenAction_POWEROFF
	}

	if responseMessage, e := client.PowerOnOrOff(ctx, &payload); e != nil {
		fmt.Println("failed to send command to unikiosk")
		return nil
	} else {
		fmt.Println("Update sent...")
		fmt.Println(responseMessage)
		fmt.Println("==============")
	}
	return nil
}

func screenshot(ctx context.Context, client service.KioskServiceClient, c config) error {
	payload := models.KioskRequest{
		Action: models.EnumScreenAction_SCREENSHOT,
	}
	screenClient, err := client.Screenshot(ctx, &payload, grpc.WaitForReady(true))
	if err != nil {
		fmt.Println("failed to send command to unikiosk", err.Error())
		return err
	}

	imageData := bytes.Buffer{}
	imageSize := 0
	done := make(chan bool)

	go func() {
		for {
			chunk, err := screenClient.Recv()
			if err != nil {
				if err == io.EOF {
					close(done)
					return
				}
				log.Println("failed to receive chunk", err.Error())
				return
			}

			size := len(chunk.Screenshot.Screenshot)
			log.Printf("received a chunk with size: %d", size)

			imageSize += size
			if imageSize > MEGABYTE*10 {
				log.Printf("image is too large: %d > %d \n", imageSize, maxImageSize)
				return
			}
			wrote, err := imageData.Write(chunk.Screenshot.Screenshot)
			if err != nil {
				log.Println("failed to write chunk", err.Error())
				return
			}
			log.Printf("wrote a chunk with size: %d", wrote)
		}
	}()

	<-done
	if c.file == "" {
		c.file = "screenshot.png"
	}

	return ioutil.WriteFile(c.file, imageData.Bytes(), 0644)
}
