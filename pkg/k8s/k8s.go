package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/darxkies/k8s-tew/pkg/config"
	"github.com/darxkies/k8s-tew/pkg/utils"
	jsonpatch "github.com/evanphx/json-patch"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

type K8S struct {
	config *config.InternalConfig
}

func NewK8S(config *config.InternalConfig) *K8S {
	return &K8S{config: config}
}

func (k8s *K8S) getConfig() (*rest.Config, error) {
	kubeconfig := k8s.config.GetFullLocalAssetFilename(utils.KubeconfigAdmin)

	// Configure connection
	config, error := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if error != nil {
		return nil, errors.Wrap(error, "Could not get Kubernetes config from flags")
	}

	return config, nil
}

func (k8s *K8S) getClient() (*kubernetes.Clientset, error) {
	// Configure connection
	config, _error := k8s.getConfig()
	if _error != nil {
		return nil, _error
	}

	// Create client
	result, _error := kubernetes.NewForConfig(config)
	if _error != nil {
		return nil, errors.Wrap(_error, "Could not get Kubernetes config")
	}

	return result, nil
}

func (k8s *K8S) unschedulable(name string, unschedulable bool) error {
	// Create client
	clientset, error := k8s.getClient()
	if error != nil {
		return error
	}

	context := context.Background()

	// Get Node
	node, error := clientset.CoreV1().Nodes().Get(context, name, metav1.GetOptions{})
	if error != nil {
		return errors.Wrapf(error, "Could not get Kubernetes node '%s'", name)
	}

	if node.Spec.Unschedulable == unschedulable {
		return nil
	}

	node.Spec.Unschedulable = unschedulable

	_, error = clientset.CoreV1().Nodes().Update(context, node, metav1.UpdateOptions{})

	if error != nil {
		return errors.Wrapf(error, "Could not update node '%s'", name)
	}

	return nil
}

func (k8s *K8S) Cordon(name string) error {
	return k8s.unschedulable(name, true)
}

func (k8s *K8S) Uncordon(name string) error {
	return k8s.unschedulable(name, false)
}

func (k8s *K8S) DeleteJob(namespace, jobName string) error {
	clientset, _error := k8s.getClient()
	if _error != nil {
		return errors.Wrapf(_error, "Could not connect to cluster")
	}

	context := context.Background()

	delete := metav1.DeletePropagationBackground

	_error = clientset.BatchV1().Jobs(namespace).Delete(context, jobName, metav1.DeleteOptions{PropagationPolicy: &delete})
	if _error != nil {
		log.WithFields(log.Fields{"error": _error, "namespace": namespace, "job": jobName}).Debug("Could not delete job")

		return _error
	}

	return nil
}

func (k8s *K8S) Drain(nodeName string) error {
	var clientset *kubernetes.Clientset
	var pods *v1.PodList

	clientset, _error := k8s.getClient()
	if _error != nil {
		return errors.Wrapf(_error, "Could not connect to cluster")
	}

	context := context.Background()
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)

	// Delete non-essential pods
	{
		log.Debug("Deleting pods")

		labelSelector := "cluster-relevant!=true"

		for {
			pods, _error = clientset.CoreV1().Pods("").List(context, metav1.ListOptions{FieldSelector: fieldSelector, LabelSelector: labelSelector})
			if _error != nil {
				return errors.Wrap(_error, "Could not get pods")
			}

			killed := 0

			for _, pod := range pods.Items {
				skip := false

				for _, ownerReference := range pod.OwnerReferences {
					if ownerReference.Kind == "Node" || ownerReference.Kind == "DaemonSet" {
						skip = true

						break
					}
				}

				for _, toleration := range pod.Spec.Tolerations {
					if toleration.Effect == v1.TaintEffectNoSchedule && toleration.Operator == v1.TolerationOpExists {
						skip = true

						break
					}
				}

				if skip {
					log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name}).Debug("Ignoring pod")

					continue
				}

				log.WithFields(log.Fields{"namespace": pod.Namespace, "pod": pod.Name}).Debug("Evicting pod")

				gracePeriodSeconds := int64(k8s.config.Config.DrainGracePeriodSeconds)

				_error := clientset.CoreV1().Pods(pod.Namespace).Delete(context, pod.Name, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds})
				if _error != nil {
					log.WithFields(log.Fields{"error": _error, "namespace": pod.Namespace, "pod": pod.Name}).Debug("Could not evict pod")
				}

				killed++
			}

			if killed == 0 {
				break
			}

			time.Sleep(time.Second)
		}
	}

	// Retrieve RBD Plugin pod
	{
		log.Debug("Looking for RBD Plugin Pod")

		done := false

		for i := 0; i < 5; i++ {
			pods, _error = clientset.CoreV1().Pods("").List(context, metav1.ListOptions{FieldSelector: fieldSelector})
			if _error != nil {
				return errors.Wrap(_error, "Could not get pods")
			}

			for _, pod := range pods.Items {
				if pod.Namespace == utils.NamespaceStorage && strings.HasPrefix(pod.Name, "csi-rbdplugin-") {
					log.WithFields(log.Fields{"pod": pod.Name}).Debug("Found RBD Plugin Pod")

					output, _error := k8s.Exec(utils.NamespaceStorage, pod.Name, "csi-rbdplugin", "for i in $(rbd device list | grep cephrbd | cut -d'/' -f 3); do rbd device unmap /dev/$i; done")
					if _error != nil {
						log.WithFields(log.Fields{"error": _error, "output": output, "pod": pod.Name, "namespace": utils.NamespaceStorage}).Debug("RBD Plugin command failed")

					} else {
						log.WithFields(log.Fields{"output": output}).Debug("RBD Plugin command successful")

						done = true

						break
					}
				}
			}

			if done {
				break
			}
		}

		if done {
			log.Debug("Removing VolumeAttachments")

			attachments, _error := clientset.StorageV1().VolumeAttachments().List(context, metav1.ListOptions{})
			if _error != nil {
				return errors.Wrap(_error, "Could not get VolumeAttachments")
			}

			for _, attachment := range attachments.Items {
				if attachment.Spec.NodeName != nodeName {
					continue
				}

				_error := clientset.StorageV1().VolumeAttachments().Delete(context, attachment.Name, metav1.DeleteOptions{})
				if _error != nil {
					log.WithFields(log.Fields{"error": _error, "namespace": attachment.Namespace, "attachment": attachment.Name}).Debug("Could not remove attachment")
				}
			}
		}
	}

	// Unmount dangling CSI mounting points
	{
		log.Debug("Looking for CSI mounts")

		for i := 0; i < int(k8s.config.Config.DrainGracePeriodSeconds); i++ {
			list := utils.GetCSIGlobalMounts(k8s.config.GetFullLocalAssetDirectory(utils.DirectoryKubeletPlugins))

			if len(list) == 0 {
				break
			}

			log.WithFields(log.Fields{"mounts": list}).Debug("Found CSI mounts")

			time.Sleep(time.Second)
		}
	}

	return nil
}

func (k8s *K8S) TaintNode(name string, nodeData *config.Node) error {
	// Create client
	clientset, error := k8s.getClient()
	if error != nil {
		return error
	}

	context := context.Background()

	// Get Node
	node, error := clientset.CoreV1().Nodes().Get(context, name, metav1.GetOptions{})
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

	node, error = clientset.CoreV1().Nodes().Update(context, node, metav1.UpdateOptions{})

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

	context := context.Background()

	secrets, error := clientset.CoreV1().Secrets(namespace).List(context, metav1.ListOptions{})
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

		isAPIService := false

		if kind == "APIService" {
			isAPIService = true
		}

		helper := resource.NewHelper(info.Client, info.Mapping)

		if !isAPIService {
			object, error = helper.Patch(
				info.Namespace,
				info.Name,
				types.ApplyPatchType,
				data,
				&options,
			)
		}

		if !isAPIService && error == nil {
			log.WithFields(log.Fields{"namespace": info.Namespace, "object": info.Name, "kind": kind, "index": i, "count": count}).Debug("Object updated")

		} else {
			if error != nil {
				log.WithFields(log.Fields{"namespace": info.Namespace, "object": info.Name, "kind": kind, "index": i, "count": count, "error": error}).Debug("Patch failed")
			}

			existingObject, error := resource.NewHelper(info.Client, info.Mapping).Get(info.Namespace, info.Name)
			if error != nil {
				object, error = resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object)
				if error != nil {
					return errors.Wrapf(error, "Could not create '%s/%s'", info.Namespace, info.Name)
				}

				log.WithFields(log.Fields{"namespace": info.Namespace, "object": info.Name, "kind": kind, "index": i, "count": count}).Debug("Object created")

			} else {
				accessor, error := meta.Accessor(existingObject)
				if error != nil {
					return errors.Wrapf(error, "Could not marshal existing '%s/%s'", info.Namespace, info.Name)
				}
				accessor.SetResourceVersion("")

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

	context := context.Background()

	secrets, error := clientset.CoreV1().Secrets(namespace).Get(context, name, metav1.GetOptions{})
	if error != nil {
		return "", "", errors.Wrapf(error, "Could not list secrets for namespace '%s'", namespace)
	}

	data, ok := secrets.Data[utils.KeyUsername]
	if !ok {
		return "", "", fmt.Errorf("Could not get username for %s/%s", namespace, name)
	}

	username = string(data)

	data, ok = secrets.Data[utils.KeyPassword]
	if !ok {
		return "", "", fmt.Errorf("Could not get password for %s/%s", namespace, name)
	}

	password = string(data)

	return username, password, nil
}

func (k8s *K8S) Exec(namespace, pod, container, command string) (string, error) {
	config, _error := k8s.getConfig()
	if _error != nil {
		return "", _error
	}

	clientset, _error := k8s.getClient()
	if _error != nil {
		return "", _error
	}

	options := &v1.PodExecOptions{
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
		Container: container,
		Command:   []string{"sh", "-c", command},
	}

	request := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(options, scheme.ParameterCodec)

	exec, _error := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if _error != nil {
		return "", errors.Wrap(_error, "Could not create k8s executor")
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	_error = exec.Stream(remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
	if _error != nil {
		return "", errors.Wrap(_error, "Could not open stream")
	}

	output := strings.TrimSpace(stdout.String()) + "\n" + strings.TrimSpace(stderr.String())

	return output, nil
}

func (k8s *K8S) WaitForCluster(totalStableIterations uint) error {
	log.Info("Waiting for Pods")

	checkPods := func() (total int, notReady int, emptyNamespaces int, _error error) {
		var clientset *kubernetes.Clientset
		var pods *v1.PodList
		var namespaces *v1.NamespaceList

		clientset, _error = k8s.getClient()

		if _error != nil {
			return
		}

		context := context.Background()

		namespaces, _error = clientset.CoreV1().Namespaces().List(context, metav1.ListOptions{})
		if _error != nil {
			return
		}

		clusterNamespaces := []string{utils.NamespaceKubeSystem, utils.NamespaceNetworking, utils.NamespaceStorage, utils.NamespaceBackup, utils.NamespaceLogging, utils.NamespaceMonitoring, utils.NamespaceShowcase}

		for _, namespace := range namespaces.Items {
			isRelevant := false

			for _, clusterNamespace := range clusterNamespaces {
				if clusterNamespace == namespace.Name {
					isRelevant = true

					break
				}
			}

			if !isRelevant {
				continue
			}

			pods, _error = clientset.CoreV1().Pods(namespace.Name).List(context, metav1.ListOptions{})
			if _error != nil {
				return
			}

			namespacePods := 0

			for _, pod := range pods.Items {
				podReady := true

				for _, container := range pod.Status.InitContainerStatuses {
					if !container.Ready {
						podReady = false
					}
				}

				for _, container := range pod.Status.ContainerStatuses {
					if !container.Ready {
						podReady = false
					}
				}

				if pod.Status.Phase != v1.PodSucceeded && !podReady {
					notReady++
				}

				total++
				namespacePods++
			}

			if namespacePods == 0 {
				emptyNamespaces++
			}
		}

		return
	}

	var lastPodsTotal int
	var stableIterations int

	for {
		podsTotal, podsNotReady, emptyNamespaces, _error := checkPods()

		if emptyNamespaces == 0 && podsTotal > 0 && lastPodsTotal == podsTotal && podsNotReady == 0 && _error == nil {
			stableIterations++

		} else {
			stableIterations = 0
		}

		lastPodsTotal = podsTotal

		if uint(stableIterations) >= totalStableIterations {
			log.WithFields(log.Fields{"pods-total": podsTotal}).Debug("Ready")

			break
		}

		log.WithFields(log.Fields{"empty-namespaces": emptyNamespaces, "pods-not-ready": podsNotReady, "pods-total": podsTotal, "error": _error, "stable-iterations": stableIterations}).Debug("Not ready")

		time.Sleep(time.Second)
	}

	return nil
}

func ApplyManifest(_config *config.InternalConfig, name, manifest string, commandRetries int) error {
	var error error

	log.WithFields(log.Fields{"name": name, "_manifest": manifest}).Info("Applying manifest")

	kubernetesClient := NewK8S(_config)

	var retries int

	for {
		if error = kubernetesClient.Apply(manifest); error == nil {
			break
		}

		log.WithFields(log.Fields{"name": name, "manifest": manifest, "error": error}).Debug("Manifest failed")

		time.Sleep(time.Second)

		retries++

		if commandRetries >= 0 && retries > commandRetries {
			break
		}
	}

	if error != nil {
		log.WithFields(log.Fields{"name": name, "manifest": manifest, "error": error}).Error("Manifest failed")

		return error
	}

	return nil
}
