// Command golangci-lint runs golangci-lint with the module plugins.
package main

import (
	"log/slog"
	"os"

	"github.com/golangci/golangci-lint/v2/pkg/commands"
	"github.com/golangci/golangci-lint/v2/pkg/exitcodes"
)

// An empty or "(devel)" version makes golangci-lint salt its analysis cache with the executable hash,
// so an edited plugin invalidates cached results.
const version = "(devel)"

func main() {
	if err := commands.Execute(commands.BuildInfo{
		Version: version,
		Date:    "(unknown)",
	}); err != nil {
		slog.Error("golangci-lint failed", "error", err)
		os.Exit(exitcodes.Failure)
	}
}
