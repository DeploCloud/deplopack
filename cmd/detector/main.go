package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/internal/deplopack"
	"github.com/railwayapp/railpack/internal/detection"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 1 {
		return fmt.Errorf("detector accepts configuration through environment variables only")
	}
	source, err := app.NewApp(".")
	if err != nil {
		return err
	}
	env := deplopack.FromEnviron(os.Environ())
	result, analyzeErr := detection.Analyze(source, env, nil)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		return err
	}
	if analyzeErr != nil {
		return analyzeErr
	}
	if !result.Success {
		return fmt.Errorf("repository analysis incomplete; see result logs")
	}
	return nil
}
