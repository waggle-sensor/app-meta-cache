package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	kube "github.com/waggle-sensor/app-meta-cache/pkg/kube_client"
	redis "github.com/waggle-sensor/app-meta-cache/pkg/redis_client"
)

var (
	WESAppCacheHost string
)

func getenv(key string, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func init() {
	rootCmd.AddCommand(verCmd)
	// To prevent printing the usage when commands end with an error
	// rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().StringVar(&WESAppCacheHost, "host", "wes-app-meta-cache", "WES app meta cache host")
}

var rootCmd = &cobra.Command{
	Use: "update-app-cache [flags] [COMMANDS]",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Waggle application meta information caching client\n")
		fmt.Printf("update-app-cache --help for more information\n")
	},
}

var verCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Redis / Kubernetes client version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Redis Client Version: %s\n", redis.REDIS_CLIENT_VER)
		fmt.Printf("Kubernetes Client Version: %s\n", kube.KUBE_CLIENT_VER)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
