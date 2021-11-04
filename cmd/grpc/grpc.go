package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mjudeikis/unikiosk/pkg/grpc/models"
	"github.com/mjudeikis/unikiosk/pkg/grpc/service"
	"google.golang.org/grpc"
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

	serverAddress := "localhost:7000"
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := service.NewKioskServiceClient(conn)

	//md := metadata.Pairs("token", "valid-token")
	//ctx = metadata.NewOutgoingContext(ctx, md)

	kioskState := models.KioskState{
		Url: "https://delfi.lt/",
	}

	if responseMessage, e := client.Start(ctx, &kioskState); e != nil {
		panic(fmt.Sprintf("Was not able to send update %v", e))
	} else {
		fmt.Println("Update sent..")
		fmt.Println(responseMessage)
		fmt.Println("=============================")
	}

	return nil
}
