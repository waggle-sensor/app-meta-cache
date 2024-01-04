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
		if err := json.Unmarshal([]byte(value), &meta); err != nil {
			return fmt.Errorf("failed to parse value %q: %s", value, err.Error())
		}
		kubeClient, err := kube.NewKubeClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to get Kubernetes client: %s", err.Error())
		}
		if nodeName == "" {
			return fmt.Errorf("no Kubernetes node name is given")
		}

		// get node labels from kubernetes
		labels, err := getNodeLabelsWithRetry(kubeClient)
		if err != nil {
			return fmt.Errorf("failed to complete the task: %s", err.Error())
		}

		// add zone label to meta and fail if it doesn't exist
		if zone, found := labels["zone"]; found {
			meta["zone"] = zone
			fmt.Printf("added zone %s\n", zone)
		} else {
			return fmt.Errorf("failed to find zone label from node %q: %v", nodeName, labels)
		}

		updatedMetaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to serialize app meta %v: %s", meta, err.Error())
		}

		// update app meta cache data
		if err := updateMetaCacheWithRetry(key, string(updatedMetaJSON)); err != nil {
			return fmt.Errorf("failed to complete the task: %s", err.Error())
		}

		return nil
	},
}

func getNodeLabelsWithRetry(kubeClient *kube.KubeClient) (map[string]string, error) {
	for i := 0; i <= retryCount; i++ {
		labels, err := kubeClient.GetNodeLabels(nodeName)
		if err != nil {
			fmt.Printf("failed to get labels from node %q: %s", nodeName, err.Error())
			fmt.Printf("%d/%d retrying in 2 seconds", i, retryCount)
			time.Sleep(2 * time.Second)
			continue
		}
		return labels, nil
	}
	return nil, fmt.Errorf("failed to get node labels")
}

func updateMetaCacheWithRetry(key, value string) error {
	redisClient := redis.NewRedisClient(WESAppCacheHost)
	for i := 0; i <= retryCount; i++ {
		if err := redisClient.Set(key, value); err != nil {
			fmt.Printf("failed to update meta at %s: %s", WESAppCacheHost, err.Error())
			fmt.Printf("%d/%d retrying in 2 seconds", i, retryCount)
			time.Sleep(2 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to update meta cache at %s", WESAppCacheHost)
}
