package cli

import (
	"errors"
	"flag"
	"fmt"
)

var ErrNoParams = errors.New("no params were informed")

type Args struct {
	Help     bool
	FilePath string
	Port     int16
}

func NewArgs() (*Args, error) {
	port := flag.Int("port", 8080, "Port to run the server on")
	help := flag.Bool("help", false, "Display this help section")
	flag.Parse()

	if *help {
		return &Args{Help: true}, nil
	}

	params := flag.Args()
	if len(params) < 1 {
		return nil, ErrNoParams
	}

	filePath := params[0]
	return &Args{FilePath: filePath, Port: int16(*port)}, nil
}

func PrintHelp() {
	fmt.Println("usage: fakeapi [FILE] [FLAGS]")
	fmt.Println("fakeapi is a fully customizable local REST API for testing.\nDetails at https://github.com/vncsmyrnk/fakeapi")
	fmt.Println("\nflags avaiable:")
	flag.PrintDefaults()
}
