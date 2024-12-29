package main

import (
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"os"

	"fakeapi/pkg/route"
	"fakeapi/pkg/server"
)

type cliArgs struct {
	FilePath string
	Port     int16
}

func newCLIArgs() (*cliArgs, error) {
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	params := flag.Args()
	if len(params) < 1 {
		return nil, errors.New("Inform the needed parameters")
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filePath := fmt.Sprintf("%s/%s", dir, params[0])

	return &cliArgs{FilePath: filePath, Port: int16(*port)}, nil
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
		server.WithPort(cliArgs.Port),
		server.WithEndpoints(endpoints),
	)

	err = server.Start()
	if err != nil {
		stdlog.Fatal(err)
	}
}
