package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	kube "github.com/waggle-sensor/app-meta-cache/pkg/kube_client"
	redis "github.com/waggle-sensor/app-meta-cache/pkg/redis_client"
)

var (
	nodeName   string
	kubeconfig string
)

func init() {
	setCmd.Flags().StringVar(&kubeconfig, "kubeconfig", getenv("KUBECONFIG", ""), "path to the kubeconfig file")
	setCmd.Flags().StringVar(&nodeName, "nodename", getenv("HOST", ""), "Kubernetes node name to get node labels")
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
		if kubeClient, err := kube.NewKubeClient(kubeconfig); err != nil {
			fmt.Printf("failed to get Kubernetes client: %s\n", err.Error())
			fmt.Printf("skip adding node label to app meta cache\n")
		} else {
			if nodeName == "" {
				fmt.Printf("no Kubernetes node name is given\n")
				fmt.Printf("skip adding node label to app meta cache\n")
			} else {
				if labels, err := kubeClient.GetNodeLabels(nodeName); err != nil {
					fmt.Printf("failed to get labels from node %q: %s\n", nodeName, err.Error())
					fmt.Printf("skip adding node label to app meta cache\n")
				} else {
					if zone, found := labels["zone"]; found {
						meta["zone"] = zone
						fmt.Printf("added zone %s\n", zone)
					} else {
						fmt.Printf("failed to find zone label from node %q: %v\n", nodeName, labels)
						fmt.Printf("skip adding node label to app meta cache\n")
					}
				}
			}
		}
		if updatedMeta, err := json.Marshal(meta); err != nil {
			return fmt.Errorf("failed to serialize app meta %v: %s", meta, err.Error())
		} else {
			redisClient := redis.NewRedisClient(WESAppCacheHost)
			return redisClient.Set(key, string(updatedMeta))
		}
	},
}
