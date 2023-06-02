package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	kube "github.com/waggle-sensor/app-meta-cache/pkg/kube_client"
	redis "github.com/waggle-sensor/app-meta-cache/pkg/redis_client"
)

var (
	nodeName   string
	kubeconfig string
	retryCount int
)

func init() {
	setCmd.Flags().StringVar(&kubeconfig, "kubeconfig", getenv("KUBECONFIG", ""), "path to the kubeconfig file")
	setCmd.Flags().StringVar(&nodeName, "nodename", getenv("HOST", ""), "Kubernetes node name to get node labels")
	setCmd.Flags().IntVar(&retryCount, "retry-count", 0, "Retry count when fails")
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set [flags] [KEY] [VALUE]",
	Short: "set KEY and VALUE in Redis server",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		meta := make(map[string]string)
		err := json.Unmarshal([]byte(value), &meta)
		if err != nil {
			return fmt.Errorf("failed to parse value %q: %s", value, err.Error())
		}
		kubeClient, err := kube.NewKubeClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to get Kubernetes client: %s", err.Error())
		}
		if nodeName == "" {
			return fmt.Errorf("no Kubernetes node name is given")
		}
		for i := 0; i <= retryCount; i++ {
			if labels, err := kubeClient.GetNodeLabels(nodeName); err != nil {
				fmt.Printf("failed to get labels from node %q: %s", nodeName, err.Error())
				fmt.Printf("%d/%d retrying in 2 seconds", i, retryCount)
				time.Sleep(2)
				continue
			} else {
				if zone, found := labels["zone"]; found {
					meta["zone"] = zone
					fmt.Printf("added zone %s\n", zone)
				} else {
					return fmt.Errorf("failed to find zone label from node %q: %v", nodeName, labels)
				}
			}
		}

		updatedMeta, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to serialize app meta %v: %s", meta, err.Error())
		}
		redisClient := redis.NewRedisClient(WESAppCacheHost)
		for i := 0; i <= retryCount; i++ {
			if err = redisClient.Set(key, string(updatedMeta)); err != nil {
				fmt.Printf("failed to set meta to %s: %s", WESAppCacheHost, err.Error())
				fmt.Printf("%d/%d retrying in 2 seconds", i, retryCount)
				time.Sleep(2)
				continue
			} else {
				return nil
			}
		}
		return fmt.Errorf("failed to complete the task: %s", err.Error())
	},
}
