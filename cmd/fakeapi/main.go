package main

import (
	"fmt"
	"log"

	args "fakeapi/internal/cli"
	"fakeapi/pkg/route"
	"fakeapi/pkg/server"
)

func main() {
	setupLog()
	cliArgs, err := args.NewCLIArgs()
	if err != nil {
		handleCLIErrors(err)
		return
	}

	if cliArgs.Help {
		args.PrintHelp()
		return
	}

	endpoints, err := route.NewEndpointsFromFile(cliArgs.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	server := server.NewServer(
		server.WithPort(cliArgs.Port),
		server.WithEndpoints(endpoints),
	)

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func handleCLIErrors(err error) {
	if err == nil {
		return
	}

	fmt.Println(err.Error())

	if err == args.ErrNoParams {
		args.PrintHelp()
		return
	}

	log.Fatal(err)
}
