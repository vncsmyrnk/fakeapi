package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var ErrNoParams = errors.New("no params were informed")

type cliArgs struct {
	FilePath string
	Port     int16
}

func newCLIArgs() (*cliArgs, error) {
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	params := flag.Args()
	if len(params) < 1 {
		return nil, ErrNoParams
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filePath := fmt.Sprintf("%s/%s", dir, params[0])

	return &cliArgs{FilePath: filePath, Port: int16(*port)}, nil
}

func printHelp() {
	fmt.Println("usage: fakeapi [FILE] [FLAGS]")
	fmt.Println("fakeapi is a fully customizable local REST API for testing.\nDetails at https://github.com/vncsmyrnk/fakeapi")
	fmt.Println("\nflags avaiable:")
	flag.PrintDefaults()
}
