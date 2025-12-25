package cli

import (
	"errors"
	"flag"
	"fmt"

	"fakeapi/internal/version"
)

var ErrNoParams = errors.New("no params were informed")

type Args struct {
	Help                    bool
	FilePath                string
	Port                    int16
	PossibleContentFilePath string
	Version                 bool
}

func NewArgs() (*Args, error) {
	port := flag.Int("port", 8080, "Port to run the server on")
	data := flag.String("data", "", "Extra content to served as content for endpoints")
	help := flag.Bool("help", false, "Display this help section")
	version := flag.Bool("version", false, "Display the app version")
	flag.Parse()

	args := &Args{
		Help:    *help,
		Version: *version,
	}

	if *help || *version {
		return args, nil
	}

	params := flag.Args()
	if len(params) < 1 {
		return nil, ErrNoParams
	}

	filePath := params[0]
	args.FilePath = filePath
	args.Port = int16(*port)
	args.PossibleContentFilePath = *data
	return args, nil
}

func PrintHelp() {
	fmt.Println("usage: fakeapi [FILE] [FLAGS]")
	fmt.Println("fakeapi is a fully customizable local REST API for testing.\nDetails at https://github.com/vncsmyrnk/fakeapi")
	fmt.Println("\nflags avaiable:")
	flag.PrintDefaults()
}

func PrintVersion() {
	fmt.Printf("%s\n", version.Version)
}
