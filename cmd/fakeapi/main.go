package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fakeapi/internal/cli"
	internallog "fakeapi/internal/log"
	"fakeapi/pkg/route"
	"fakeapi/pkg/server"
)

func main() {
	internallog.SetupLog()
	cliArgs, err := cli.NewArgs()
	if err != nil {
		handleCLIErrors(err)
		return
	}

	if cliArgs.Help {
		cli.PrintHelp()
		return
	}

	if cliArgs.Version {
		cli.PrintVersion()
		return
	}

	endpoints, err := route.NewEndpointsFromFile(cliArgs.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	statePath, err := getStatePath()
	if err != nil {
		log.Fatal(err)
	}

	server := server.NewServer(
		server.WithPort(cliArgs.Port),
		server.WithEndpoints(endpoints),
		server.WithEndpointsPossibleContentFilePath(cliArgs.PossibleContentFilePath),
		server.WithStatePath(statePath),
	)

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func getStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "state", "fakeapi"), nil
}

func handleCLIErrors(err error) {
	if err == nil {
		return
	}

	fmt.Println(err.Error())

	if err == cli.ErrNoParams {
		cli.PrintHelp()
		return
	}

	log.Fatal(err)
}
