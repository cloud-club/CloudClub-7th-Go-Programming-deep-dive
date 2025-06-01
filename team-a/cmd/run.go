package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"swarm/internal/model"
	"sync"
	"time"
)

var (
	host     string
	users    int
	duration time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a simple load test",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting test: %d users, target=%s, duration=%s\n", users, host, duration)

		results := make(chan model.Result, users*100) // 충분히 큼직하게

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		for i := 0; i < users; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &http.Client{}
				for {
					select {
					case <-ctx.Done():
						return
					default:
						start := time.Now()
						resp, err := client.Get(host)
						elapsed := time.Since(start)
						if err != nil {

							results <- model.Result{UserID: id, Duration: elapsed, Timestamp: time.Now(), Error: err.Error()}
						} else {
							results <- model.Result{UserID: id, StatusCode: resp.StatusCode, Duration: elapsed, Timestamp: time.Now()}
							resp.Body.Close()
						}
						time.Sleep(500 * time.Millisecond)
					}
				}
			}(i)
		}

		wg.Wait()
		fmt.Println("Load test completed")

		close(results)

		var allResults []model.Result
		for res := range results {
			allResults = append(allResults, res)
		}

		f, err := os.Create("results.json")
		if err != nil {
			fmt.Println("Failed to create file:", err)
			return
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				fmt.Println("Failed to close file:", err)
			}
		}(f)

		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ") // pretty print
		if err := encoder.Encode(allResults); err != nil {
			fmt.Println("Failed to write JSON:", err)
		}
	},
}

func init() {
	runCmd.Flags().StringVarP(&host, "host", "H", "http://localhost", "Target host URL")
	runCmd.Flags().IntVarP(&users, "users", "u", 1, "Number of concurrent users")
	runCmd.Flags().DurationVarP(&duration, "duration", "d", 10*time.Second, "Test duration (e.g., 10s, 1m)")
	rootCmd.AddCommand(runCmd)
}
