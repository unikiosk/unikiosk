package set

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"

	"github.com/unikiosk/unikiosk/pkg/api"
)

type config struct {
	url               string
	unikioskServerUrl string
	file              string
	action            string
	screenResolution  string
}

// New returns the cobra command for "set".
func New() *cobra.Command {
	var c config
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Change screen configuration",
		Long:  "Manage screen configuration ",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validate(c)
			if err != nil {
				return err
			}

			return set(cmd.Context(), c)
		},
	}

	cmd.Flags().StringVarP(&c.url, "url", "u", "", "Set desired URL to be opened")
	cmd.Flags().StringVarP(&c.unikioskServerUrl, "server", "s", "http://localhost:8081/api", "Set desired URL to be opened")
	cmd.Flags().StringVarP(&c.file, "file", "f", "", "File to write screenshoot")
	cmd.Flags().StringVarP(&c.action, "action", "a", "update", "Screen action [start,update,poweron,poweroff]")
	cmd.Flags().StringVarP(&c.screenResolution, "resolution", "", "", "Screen resolution [widthXheight]. Example: 1920x1080")

	return cmd
}

func set(ctx context.Context, c config) error {
	req := api.KioskRequest{}
	if c.url != "" {
		req.Content = c.url
	}
	if c.file != "" {
		req.Content = c.file
	}
	// set size
	res := strings.ToLower(c.screenResolution)
	if c.screenResolution != "" && strings.Contains(res, "x") {
		parts := strings.Split(res, "x")
		w, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid resolution: %s", c.screenResolution)
		}
		h, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid resolution: %s", c.screenResolution)
		}
		req.SizeW = w
		req.SizeH = h
	}

	action, err := api.StringToAction(c.action)
	if err != nil {
		return err
	}
	req.Action = action

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal: %s", err.Error())
	}

	response, err := http.Post(c.unikioskServerUrl, api.ContentTypeApplicationJSON, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to update screen while reaching out to screen: %s", err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("screen is not reachable: %d", response.StatusCode)
	}

	result := api.KioskResponse{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	spew.Dump(result)
	return nil
}

func validate(c config) error {
	if c.action != "update" {
		if c.file != "" || c.url != "" {
			return fmt.Errorf("--action can't be used with --url or --file")
		}
	}
	//if c.file != "" && c.url != "" {
	//	return fmt.Errorf("only --url or --file can be provided")
	//}
	//if c.file == "" && c.url == "" && c.action == "update" {
	//	return fmt.Errorf("one of --url or --file must be provided")
	//}

	// TODO: validate url and file format/content
	return nil

}
