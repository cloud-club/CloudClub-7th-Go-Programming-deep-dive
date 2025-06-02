package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sort"
	"swarm/internal/model"
)

var inputFile string

var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "analysis command",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(inputFile)
		if err != nil {
			log.Fatal(err)
			return
		}

		defer f.Close()

		var results []model.Result
		if err := json.NewDecoder(f).Decode(&results); err != nil {
			fmt.Println("Failed to parse JSON:", err)
			return
		}

		total := len(results)
		var success, fail int
		var totalDuration int64
		var durations []int64

		for _, result := range results {
			if result.Error != "" || result.StatusCode >= 400 {
				fail++
			} else {
				success++
			}
			durations = append(durations, result.Duration)
			totalDuration += result.Duration
		}

		sort.Slice(durations, func(i int, j int) bool {
			return durations[i] < durations[j]
		})

		// 익명함수
		percentile := func(p float64) int64 {
			if len(durations) == 0 {
				return 0
			}
			rank := int(float64(len(durations))*p + 0.5)
			if rank >= len(durations) {
				rank = len(durations) - 1
			}
			return durations[rank]
		}
		fmt.Println("📊 Load Test Analysis")
		fmt.Println("────────────────────────────")
		fmt.Printf("Total Requests:        %d\n", total)
		fmt.Printf("Successful (2xx):      %d (%.2f%%)\n", success, float64(success)/float64(total)*100)
		fmt.Printf("Failed (4xx/5xx/errors): %d (%.2f%%)\n", fail, float64(fail)/float64(total)*100)

		if total > 0 {
			fmt.Printf("\nAvg Response Time:     %d ms\n", totalDuration/int64(total))
			fmt.Printf("Min Response Time:     %d ms\n", durations[0])
			fmt.Printf("Max Response Time:     %d ms\n", durations[len(durations)-1])
			fmt.Printf("P90 Response Time:     %d ms\n", percentile(0.90))
			fmt.Printf("P95 Response Time:     %d ms\n", percentile(0.95))
			fmt.Printf("P99 Response Time:     %d ms\n", percentile(0.99))
		}

	},
}

func init() {
	analysisCmd.Flags().StringVarP(&inputFile, "input", "i", "results.json", "Path to JSON results file")
	rootCmd.AddCommand(analysisCmd)
}
