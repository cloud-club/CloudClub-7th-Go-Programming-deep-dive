/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "swarm",
	Short: "A lightweight load testing CLI built with Go",
	Long: `Swarm is a simple and extensible load testing tool written in Go.

It helps developers and QA engineers simulate concurrent HTTP requests,
measure performance metrics, and analyze results easily.

Supported features:
  - Concurrent user simulation using goroutines
  - Percentile analysis (P90, P95, P99)
  - JSON result export and console analysis

Examples:

  Run a basic load test:
    swarm run --host http://example.com --users 10 --duration 30s

  Analyze the test result:
    swarm analysis --input results.json

`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
