package k8s

import (
	"fmt"
	"strings"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8S struct {
	config *config.InternalConfig
}

func NewK8S(config *config.InternalConfig) *K8S {
	return &K8S{config: config}
}

func (k8s *K8S) getClient() (*kubernetes.Clientset, error) {
	kubeconfig := k8s.config.GetFullLocalAssetFilename(utils.KubeconfigAdmin)

	// Configure connection
	config, error := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if error != nil {
		return nil, error
	}

	// Create client
	return kubernetes.NewForConfig(config)
}

func (k8s *K8S) TaintNode(name string, nodeData *config.Node) error {
	// Create client
	clientset, error := k8s.getClient()
	if error != nil {
		return error
	}

	// Get Node
	node, error := clientset.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if error != nil {
		return error
	}

	changed := false

	addLabel := func(label string) {
		if _, ok := node.Labels[label]; !ok {
			changed = true
		}

		node.Labels[label] = "true"
	}

	removeLabel := func(label string) {
		if _, ok := node.Labels[label]; ok {
			changed = true
		}

		delete(node.Labels, label)
	}

	addTaint := func(label string) {
		found := false

		for _, taint := range node.Spec.Taints {
			if taint.Key == label {
				found = true

				break
			}
		}

		if !found {
			node.Spec.Taints = append(node.Spec.Taints, v1.Taint{Key: label, Value: "true", Effect: v1.TaintEffectNoSchedule})

			changed = true
		}

		addLabel(label)
	}

	removeTaint := func(label string) {
		taints := []v1.Taint{}

		for _, taint := range node.Spec.Taints {
			if taint.Key == label {
				changed = true

				continue
			}

			taints = append(taints, taint)
		}

		node.Spec.Taints = taints

		removeLabel(label)
	}

	if nodeData.IsControllerAndWorker() {
		addLabel(utils.ControllerOnlyTaintKey)

	} else {
		removeLabel(utils.ControllerOnlyTaintKey)
	}

	if nodeData.IsControllerOnly() {
		addTaint(utils.ControllerOnlyTaintKey)

	} else {
		removeTaint(utils.ControllerOnlyTaintKey)
	}

	if nodeData.IsStorageOnly() {
		addTaint(utils.StorageOnlyTaintKey)

	} else {
		removeTaint(utils.StorageOnlyTaintKey)
	}

	if nodeData.IsWorkerOnly() {
		addLabel(utils.WorkerOnlyTaintKey)

	} else {
		removeLabel(utils.WorkerOnlyTaintKey)
	}

	if !changed {
		return nil
	}

	_, error = clientset.CoreV1().Nodes().Update(node)

	return error
}

func (k8s *K8S) GetSecretToken(namespace, name string) (string, error) {
	clientset, error := k8s.getClient()
	if error != nil {
		return "", error
	}

	secrets, error := clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if error != nil {
		return "", error
	}

	for _, secret := range secrets.Items {
		if strings.HasPrefix(secret.Name, fmt.Sprintf("%s-token-", name)) {
			if value, ok := secret.Data["token"]; ok {
				return string(value), nil
			}
		}
	}

	return "", fmt.Errorf("No token with prefix '%s' found", name)
}
