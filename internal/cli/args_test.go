package cli

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestFlags(t *testing.T) {
	testCases := []struct {
		name         string
		cliArgs      []string
		expectedArgs *Args
		error        bool
	}{
		{
			name:    "success",
			cliArgs: []string{"cmd/fakeapi", "--port", "9090", "file.json"},
			expectedArgs: &Args{
				Port:     int16(9090),
				FilePath: "file.json",
			},
		},
		{
			name:         "file missing",
			cliArgs:      []string{"cmd/fakeapi"},
			expectedArgs: nil,
			error:        true,
		},
		{
			name:    "default port",
			cliArgs: []string{"cmd/fakeapi", "config.json"},
			expectedArgs: &Args{
				Port:     int16(8080),
				FilePath: "config.json",
			},
		},
		{
			name:         "help",
			cliArgs:      []string{"cmd/fakeapi", "--help"},
			expectedArgs: &Args{Help: true},
		},
		{
			name:         "version",
			cliArgs:      []string{"cmd/fakeapi", "--version"},
			expectedArgs: &Args{Version: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure tests are not run in parallel
			// Do not call t.Parallel() here
			resetFlags()

			os.Args = tc.cliArgs
			args, err := NewArgs()

			if tc.error {
				assert.Error(t, err)
				assert.Nil(t, args)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}
