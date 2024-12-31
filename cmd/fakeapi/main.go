package main

import (
	"fmt"
	"log"

	"fakeapi/pkg/route"
	"fakeapi/pkg/server"
)

func main() {
	setupLog()
	cliArgs, err := newCLIArgs()
	if err != nil {
		handleCLIErrors(err)
		return
	}

	if cliArgs.Help {
		printHelp()
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

	if err == ErrNoParams {
		printHelp()
		return
	}

	log.Fatal(err)
}
