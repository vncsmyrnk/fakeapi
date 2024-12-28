package main

import (
	"errors"
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

	return &cliArgs{FilePath: os.Args[1]}, nil
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

	server.Run(server.Config{
		Port:      8080,
		Endpoints: endpoints,
	})
}
