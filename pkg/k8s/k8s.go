package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
	jsonpatch "github.com/evanphx/json-patch"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
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
		return nil, errors.Wrap(error, "Could not get Kubernetes config from flags")
	}

	// Create client
	result, error := kubernetes.NewForConfig(config)
	if error != nil {
		return nil, errors.Wrap(error, "Could not get Kubernetes config")
	}

	return result, nil
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
		return errors.Wrapf(error, "Could not get Kubernetes node '%s'", name)
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
	}

	if nodeData.IsControllerOnly() {
		addTaint(utils.NodeRoleController)

	} else {
		removeTaint(utils.NodeRoleController)
	}

	if nodeData.IsStorageOnly() {
		addTaint(utils.NodeRoleStorage)

	} else {
		removeTaint(utils.NodeRoleStorage)
	}

	if nodeData.IsController() {
		addLabel(utils.NodeRoleController)

	} else {
		removeLabel(utils.NodeRoleController)
	}

	if nodeData.IsWorker() {
		addLabel(utils.NodeRoleWorker)

	} else {
		removeLabel(utils.NodeRoleWorker)
	}

	if nodeData.IsStorage() {
		addLabel(utils.NodeRoleStorage)

	} else {
		removeLabel(utils.NodeRoleStorage)
	}

	if !changed {
		return nil
	}

	node, error = clientset.CoreV1().Nodes().Update(node)

	if error != nil {
		return errors.Wrapf(error, "Could not update node '%s'", name)
	}

	return nil
}

func (k8s *K8S) GetSecretToken(namespace, name string) (string, error) {
	clientset, error := k8s.getClient()
	if error != nil {
		return "", error
	}

	secrets, error := clientset.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if error != nil {
		return "", errors.Wrapf(error, "Could not list secrets for namespace '%s'", namespace)
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

func (k8s *K8S) Apply(manifest string) error {
	kubeConfig := k8s.config.GetFullLocalAssetFilename(utils.KubeconfigAdmin)

	getter := genericclioptions.NewConfigFlags(true)

	getter.KubeConfig = &kubeConfig

	factory := cmdutil.NewFactory(getter)

	schema, error := factory.Validator(false)
	if error != nil {
		return errors.Wrapf(error, "Could not generate validator for '%s'", manifest)
	}

	filenameOptions := &resource.FilenameOptions{Recursive: true, Filenames: []string{manifest}}

	resources := factory.NewBuilder().
		ContinueOnError().
		Unstructured().
		Schema(schema).
		DefaultNamespace().
		FilenameParam(false, filenameOptions).
		Flatten().
		Do()

	if error := resources.Err(); error != nil {
		return errors.Wrapf(error, "Could not get manifest resources for '%s'", manifest)
	}

	infos, error := resources.Infos()
	if error != nil {
		error = errors.Wrapf(error, "Could not decode manifest '%s'", manifest)
	}
	count := len(infos)

	for i, info := range infos {
		var object runtime.Object

		kind := info.Mapping.GroupVersionKind.Kind

		data, error := runtime.Encode(unstructured.UnstructuredJSONScheme, info.Object)
		if error != nil {
			return errors.Wrapf(error, "Could not encode '%s/%s/%s'", info.Namespace, kind, info.Name)
		}

		force := true

		options := metav1.PatchOptions{
			Force:        &force,
			FieldManager: "kubectl",
		}

		helper := resource.NewHelper(info.Client, info.Mapping)
		object, error = helper.Patch(
			info.Namespace,
			info.Name,
			types.ApplyPatchType,
			data,
			&options,
		)
		if error == nil {
			log.WithFields(log.Fields{"namespace": info.Namespace, "object": info.Name, "kind": kind, "index": i, "count": count}).Debug("Object updated")

		} else {
			existingObject, error := resource.NewHelper(info.Client, info.Mapping).Get(info.Namespace, info.Name, true)
			if error != nil {
				object, error = resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object, nil)
				if error != nil {
					return errors.Wrapf(error, "Could not create '%s/%s'", info.Namespace, info.Name)
				}

				log.WithFields(log.Fields{"namespace": info.Namespace, "object": info.Name, "kind": kind, "index": i, "count": count}).Debug("Object created")

			} else {
				existingJson, error := json.Marshal(existingObject)
				if error != nil {
					return errors.Wrapf(error, "Could not marshal existing '%s/%s'", info.Namespace, info.Name)
				}

				targetJson, error := json.Marshal(info.Object)
				if error != nil {
					return errors.Wrapf(error, "Could not marshal target '%s/%s'", info.Namespace, info.Name)
				}

				patch, error := jsonpatch.CreateMergePatch(existingJson, targetJson)
				if error != nil {
					return errors.Wrapf(error, "Could not create patch '%s/%s'", info.Namespace, info.Name)
				}

				object, error = resource.NewHelper(info.Client, info.Mapping).Patch(info.Namespace, info.Name, types.MergePatchType, patch, nil)
				if error != nil {
					return errors.Wrapf(error, "Could not patch '%s/%s'", info.Namespace, info.Name)
				}

				log.WithFields(log.Fields{"namespace": info.Namespace, "object": info.Name, "kind": kind, "index": i, "count": count}).Debug("Object patched")
			}
		}

		if error := info.Refresh(object, true); error != nil {
			return errors.Wrapf(error, "Could not refresh '%s/%s'", info.Namespace, info.Name)
		}

	}

	return error
}

func (k8s *K8S) GetCredentials(namespace, name string) (username string, password string, error error) {
	clientset, error := k8s.getClient()
	if error != nil {
		return "", "", error
	}

	secrets, error := clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if error != nil {
		return "", "", errors.Wrapf(error, "Could not list secrets for namespace '%s'", namespace)
	}

	data, ok := secrets.Data["username"]
	if !ok {
		return "", "", fmt.Errorf("Could not get username for %s/%s", namespace, name)
	}

	username = string(data)

	data, ok = secrets.Data["password"]
	if !ok {
		return "", "", fmt.Errorf("Could not get password for %s/%s", namespace, name)
	}

	password = string(data)

	return username, password, nil
}
