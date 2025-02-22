// The three executables (check, in, out) are symlinked to this file.
// For statically linked binaries like Go, this reduces the size by 2/3.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/Pix4D/cogito/cogito"
	"github.com/Pix4D/cogito/github"
	"github.com/Pix4D/cogito/sets"
	"github.com/hashicorp/go-hclog"
)

func main() {
	// The "Concourse resource protocol" expects:
	// - stdin, stdout and command-line arguments for the protocol itself
	// - stderr for logging
	// See: https://concourse-ci.org/implementing-resource-types.html
	if err := mainErr(os.Stdin, os.Stdout, os.Stderr, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "cogito: error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr(in io.Reader, out io.Writer, logOut io.Writer, args []string) error {
	cmd := path.Base(args[0])
	validCmds := sets.From("check", "in", "out")
	if !validCmds.Contains(cmd) {
		return fmt.Errorf("invoked as '%s'; want: one of %v", cmd, validCmds)
	}

	input, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("reading stdin: %s", err)
	}
	logLevel, err := peekLogLevel(input)
	if err != nil {
		return err
	}
	log := hclog.New(&hclog.LoggerOptions{
		Name:        "cogito",
		Level:       hclog.LevelFromString(logLevel),
		Output:      logOut,
		DisableTime: true,
	})
	log.Info(cogito.BuildInfo())

	ghAPI := os.Getenv("COGITO_GITHUB_API")
	if ghAPI != "" {
		log.Info("endpoint override", "COGITO_GITHUB_API", ghAPI)
	} else {
		ghAPI = github.API
	}

	switch cmd {
	case "check":
		return cogito.Check(log, input, out, args[1:])
	case "in":
		return cogito.Get(log, input, out, args[1:])
	case "out":
		putter := cogito.NewPutter(ghAPI, log)
		return cogito.Put(log, input, out, args[1:], putter)
	default:
		return fmt.Errorf("cli wiring error; please report")
	}
}

// peekLogLevel decodes 'input' as JSON and looks for key source.log_level. If 'input'
// is not JSON, peekLogLevel will return an error. If 'input' is JSON but does not
// contain key source.log_level, peekLogLevel returns "info" as default value.
//
// Rationale: depending on the Concourse step we are invoked for, the JSON object we get
// from stdin is different, but it always contains a struct with name "source", thus we
// can peek into it to gather the log level as soon as possible.
func peekLogLevel(input []byte) (string, error) {
	type Peek struct {
		Source struct {
			LogLevel string `json:"log_level"`
		} `json:"source"`
	}
	var peek Peek
	peek.Source.LogLevel = "info" // default value
	if err := json.Unmarshal(input, &peek); err != nil {
		return "", fmt.Errorf("peeking into JSON for log_level: %s", err)
	}

	return peek.Source.LogLevel, nil
}
