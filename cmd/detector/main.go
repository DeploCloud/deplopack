package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/internal/detection"
)

var version = "dev"

func main() {
	configFile := flag.String("config-file", "", "Config file relative to the source directory")
	showVersion := flag.Bool("version", false, "Print detector version")
	var envs []string
	flag.Func("env", "Project environment variable NAME=VALUE (repeatable)", func(value string) error {
		envs = append(envs, value)
		return nil
	})
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: deplopack-detector [flags] DIRECTORY")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), envs, *configFile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(directory string, envs []string, configFile string) error {
	source, err := app.NewApp(directory)
	if err != nil {
		return err
	}
	info, err := os.Stat(source.Source)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source must be a directory")
	}
	env, err := app.FromEnvs(envs)
	if err != nil {
		return err
	}
	result, analyzeErr := detection.Analyze(source, env, &core.GenerateBuildPlanOptions{ConfigFilePath: configFile})
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
