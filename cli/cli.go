// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	terminal "golang.org/x/term"

	snapshot_conf "github.com/mendersoftware/mender-snapshot/conf"
)

const (
	snapshotDescription = "Creates a snapshot of the currently running " +
		"rootfs. The snapshots can be passed as a rootfs-image to the " +
		"mender-artifact tool to create an update based on THIS " +
		"device's rootfs. Refer to the list of COMMANDS to specify " +
		"where to stream the image.\n" +
		"\t NOTE: If the process gets killed (e.g. by SIGKILL) " +
		"while a snapshot is in progress, the system may freeze - " +
		"forcing you to manually hard-reboot the device. " +
		"Use at your own risk - preferably on a device that " +
		"is physically accessible."
	snapshotDumpDescription = "Dump rootfs to standard out. Exits if " +
		"output isn't redirected."
)

var (
	errDumpTerminal = errors.New("Refusing to write to terminal")
)

type logOptionsType struct {
	logLevel string
}

type runOptionsType struct {
	logOptions logOptionsType
}

func ShowVersion() string {
	return fmt.Sprintf("%s\truntime: %s",
		snapshot_conf.VersionString(), runtime.Version())
}

func SetupCLI(args []string) error {
	runOptions := &runOptionsType{}

	app := &cli.App{
		Name: "snapshot",
		Usage: "Create filesystem snapshot -" +
			"'mender-snapshot --help' for more.",
		ArgsUsage:   "[options]",
		Description: snapshotDescription,
		Version:     ShowVersion(),
		Commands: []*cli.Command{
			{
				Name:        "dump",
				Description: snapshotDumpDescription,
				Usage:       "Dumps rootfs to stdout.",
				Action:      runOptions.DumpSnapshot,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "data",
						Aliases: []string{"d"},
						Usage:   "Mender state data `DIR`ECTORY path.",
						Value:   snapshot_conf.DefaultDataStore,
					},
					&cli.StringFlag{
						Name: "source",
						Usage: "Path to target " +
							"filesystem " +
							"file/directory/device" +
							"to snapshot.",
						Value: "/",
					},
					&cli.BoolFlag{
						Name:    "quiet",
						Aliases: []string{"q"},
						Usage: "Suppress output " +
							"and only report " +
							"logs from " +
							"ERROR level",
					},
					&cli.StringFlag{
						Name:    "compression",
						Aliases: []string{"C"},
						Usage: "Compression type to use on the" +
							"rootfs snapshot {none,gzip}",
						Value: "none",
					},
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"l"},
			Usage:       "Set logging `level`.",
			Value:       "info",
			Destination: &runOptions.logOptions.logLevel,
		},
	}

	cli.HelpPrinter = upgradeHelpPrinter(cli.HelpPrinter)
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(c.App.Writer, "%s\n", ShowVersion())
	}
	return app.Run(args)
}

func upgradeHelpPrinter(defaultPrinter func(w io.Writer, templ string, data interface{})) func(
	w io.Writer, templ string, data interface{}) {
	// Applies the ordinary help printer with column post processing
	return func(stdout io.Writer, templ string, data interface{}) {
		// Need at least 10 characters for last column in order to
		// pretty print; otherwise the output is unreadable.
		const minColumnWidth = 10
		isLowerCase := func(c rune) bool {
			// returns true if c in [a-z] else false
			asciiVal := int(c)
			if asciiVal >= 0x61 && asciiVal <= 0x7A {
				return true
			}
			return false
		}
		// defaultPrinter parses the text-template and outputs to buffer
		var buf bytes.Buffer
		defaultPrinter(&buf, templ, data)
		terminalWidth, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil || terminalWidth <= 0 {
			// Just write help as is.
			stdout.Write(buf.Bytes())
			return
		}
		for line, err := buf.ReadString('\n'); err == nil; line, err = buf.ReadString('\n') {
			if len(line) <= terminalWidth+1 {
				stdout.Write([]byte(line))
				continue
			}
			newLine := line
			indent := strings.LastIndex(
				line[:terminalWidth], "  ")
			// find indentation of last column
			if indent == -1 {
				indent = 0
			}
			indent += strings.IndexFunc(
				strings.ToLower(line[indent:]), isLowerCase) - 1
			if indent >= terminalWidth-minColumnWidth ||
				indent == -1 {
				indent = 0
			}
			// Format the last column to be aligned
			for len(newLine) > terminalWidth {
				// find word to insert newline
				idx := strings.LastIndex(newLine[:terminalWidth], " ")
				if idx == indent || idx == -1 {
					idx = terminalWidth
				}
				stdout.Write([]byte(newLine[:idx] + "\n"))
				newLine = newLine[idx:]
				newLine = strings.Repeat(" ", indent) + newLine
			}
			stdout.Write([]byte(newLine))
		}
		if err != nil {
			log.Fatalf("CLI HELP: error writing help string: %v\n", err)
		}
	}
}
