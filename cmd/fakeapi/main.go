package main

import (
	"errors"
	"fmt"
	stdlog "log"
	"os"

	"fakeapi/pkg/route"
	"fakeapi/pkg/server"
)

type cliArgs struct {
	FilePath string
}

func newCLIArgs() (*cliArgs, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("Inform the needed parameters")
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filePath := fmt.Sprintf("%s/%s", dir, os.Args[1])

	return &cliArgs{FilePath: filePath}, nil
}

func main() {
	cliArgs, err := newCLIArgs()
	if err != nil {
		stdlog.Fatal(err)
	}

	endpoints, err := route.NewEndpointsFromFile(cliArgs.FilePath)
	if err != nil {
		stdlog.Fatal(err)
	}

	server := server.NewServer(
		server.WithPort(int16(8080)),
		server.WithEndpoints(endpoints),
	)

	err = server.Start()
	if err != nil {
		stdlog.Fatal(err)
	}
}
