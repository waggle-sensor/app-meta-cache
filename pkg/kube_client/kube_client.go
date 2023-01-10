package kube_client

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubernetes client-go version coming from go.mod
const KUBE_CLIENT_VER = "v0.26.0"

type KubeClient struct {
	client *kubernetes.Clientset
}

func NewKubeClient(configFilePath string) (*KubeClient, error) {
	incluster := true
	if len(configFilePath) > 0 {
		incluster = false
	}
	client, err := getK3SClient(incluster, configFilePath)
	if err != nil {
		return nil, err
	}
	return &KubeClient{
		client: client,
	}, nil
}

func (k *KubeClient) GetNodeLabels(nodeName string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node, err := k.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return node.Labels, nil
}

// GetK3SClient returns an instance of clientset talking to a K3S cluster
func getK3SClient(incluster bool, pathToConfig string) (*kubernetes.Clientset, error) {
	if incluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(config)
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", pathToConfig)
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(config)
	}
}
