package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/mjudeikis/unikiosk/pkg/grpc/models"
	"github.com/mjudeikis/unikiosk/pkg/grpc/service"
)

var (
	url            = flag.String("url", "https://synpse.net", `URL to open`)
	kioskServerUrl = flag.String("unikiosk-url", "localhost:7000", `URL of unikiosk instance`)
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

	kioskState := models.KioskState{
		Url: *url,
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
