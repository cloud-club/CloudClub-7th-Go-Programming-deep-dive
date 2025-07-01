package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"swarm/internal/model"
	"sync"
	"time"
)

// Parse the yaml file
type Config struct {
	Host     string        `yaml:"host"`
	Duration time.Duration `yaml:"duration"`
	Users    int           `yaml:"users"`
	Paths    []PathConfig  `yaml:"paths"`
}

type PathConfig struct {
	Path  string `yaml:"path"`
	Ratio int    `yaml:"ratio"`
}

type PathUserConfig struct {
	Path      string
	UserCount int
}

var (
	cfgFile  string
	cfg      Config
	host     string
	users    int
	duration time.Duration
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a simple load test",

	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		fmt.Printf("Configuration loaded from: %s\n", viper.ConfigFileUsed())

		finalHost := viper.GetString("host")
		finalUsers := viper.GetInt("users")
		finalDuration := viper.GetDuration("duration")

		// 커맨드라인 플래그가 기본값이 아니면 플래그 값 사용, 아니면 설정 파일 값 사용
		if !cmd.Flags().Changed("host") && cfg.Host != "" {
			host = finalHost
		}

		if !cmd.Flags().Changed("users") && cfg.Users != 0 {
			users = finalUsers
		}

		if !cmd.Flags().Changed("duration") && cfg.Duration != 0 {
			duration = finalDuration
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Start load test: %d users, target=%s, duration=%s\n", users, host, duration)

		// Calculate user ratio
		remainingUsers := users
		var distribution []PathUserConfig
		for i, path := range cfg.Paths {
			var userCount int

			if i == len(cfg.Paths)-1 {
				// 마지막 경로는 나머지 사용자 전부 할당
				userCount = remainingUsers
			} else {
				// 비율에 따른 사용자 수 계산
				userCount = (cfg.Users * path.Ratio) / 100
				remainingUsers -= userCount
			}

			if userCount > 0 {
				distribution = append(distribution, PathUserConfig{
					Path:      path.Path,
					UserCount: userCount,
				})
			}
		}
		fmt.Println(distribution)

		results := make(chan model.Result, users*100) // 충분히 큼직하게

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), duration)
		defer cancel()

		userID := 0
		for _, value := range distribution {
			for i := 0; i < value.UserCount; i++ {
				wg.Add(1)
				userID++

				go func(id int, path string) {
					defer wg.Done()
					client := &http.Client{}
					url := host + value.Path

					for {
						select {
						case <-ctx.Done():
							return
						default:
							start := time.Now()
							resp, err := client.Get(url)
							elapsed := time.Since(start).Milliseconds()
							if err != nil {
								results <- model.Result{UserID: id, Duration: elapsed, Timestamp: time.Now(), Error: err.Error()}
							} else {
								results <- model.Result{UserID: id, StatusCode: resp.StatusCode, Duration: elapsed, Timestamp: time.Now()}
								resp.Body.Close()
							}
							time.Sleep(1000 * time.Millisecond)
						}
					}
				}(userID, value.Path)
			}
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
	//cobra.OnInitialize(initConfig)
	runCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/config.yaml)")
	runCmd.Flags().StringVarP(&host, "host", "H", "http://localhost", "Target host URL")
	runCmd.Flags().IntVarP(&users, "users", "U", 1, "Number of concurrent users")
	runCmd.Flags().DurationVarP(&duration, "duration", "d", 10*time.Second, "Test duration (e.g., 10s, 1m)")
	//runCmd.Flags().Bool("viper", true, "use Viper for configuration")
	rootCmd.AddCommand(runCmd)
}

func initConfig() {
	// load home dir
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Can't find config file:", viper.ConfigFileUsed())
		os.Exit(1)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	validatePathConfig()

	fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
}

func validatePathConfig() {
	totalRatio := 0
	for _, path := range cfg.Paths {
		if path.Path == "" {
			fmt.Println("Path must be not empty")
			os.Exit(1)
		}

		totalRatio += path.Ratio
		fmt.Println("Path:", path.Path, "Ratio:", path.Ratio)
	}

	if totalRatio != 100 {
		fmt.Println("Ratio must be equal to 100")
		os.Exit(1)
	}
}
