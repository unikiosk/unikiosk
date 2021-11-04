package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/mjudeikis/unikiosk/pkg/grpc/models"
	"github.com/mjudeikis/unikiosk/pkg/grpc/service"
	fileutil "github.com/mjudeikis/unikiosk/pkg/util/file"
)

var (
	url            = flag.String("url", "https://synpse.net", `URL to open`)
	kioskServerUrl = flag.String("unikiosk-url", "localhost:7000", `URL of UniKiosk instance`)
	file           = flag.String("file", "", `relative path to html file to serve`)
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	flag.Parse()

	conn, err := grpc.Dial(*kioskServerUrl, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := service.NewKioskServiceClient(conn)

	payload := *url
	if *file != "" {
		exists, _ := fileutil.Exist(*file)
		if !exists {
			return fmt.Errorf("file [%s] not found", *file)
		}
		var err error
		data, err := ioutil.ReadFile(*file)
		if err != nil {
			return err
		}

		payload = `data:text/html,
		` + string(data) + `
		`
	}

	kioskState := models.KioskState{
		Content: payload,
	}

	if responseMessage, e := client.StartOrUpdate(ctx, &kioskState); e != nil {
		panic(fmt.Sprintf("Was not able to send update %v", e))
	} else {
		fmt.Println("Update sent...")
		fmt.Println(responseMessage)
		fmt.Println("=============================")
	}

	return nil
}
