package config

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"

	"github.com/darxkies/k8s-tew/utils"

	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v2"
)

type Command struct {
	Command string
	Labels  Labels
}

func NewCommand(labels Labels, command string) *Command {
	return &Command{Labels: labels, Command: command}
}

type Labels []string

func (labels Labels) HasLabels(otherLabels Labels) bool {
	for _, label := range labels {
		for _, otherLabel := range otherLabels {
			if label == otherLabel {
				return true
			}
		}
	}

	return false
}

type DeploymentFile struct {
	Labels Labels `yaml:"labels,omitempty"`
	File   string `yaml:"file"`
}

func NewDeploymentFile(labels []string, file string) *DeploymentFile {
	return &DeploymentFile{Labels: labels, File: file}
}

type ServerConfig struct {
	Labels    Labels            `yaml:"labels"`
	Command   string            `yaml:"command"`
	Arguments map[string]string `yaml:"arguments"`
}

func (config ServerConfig) Dump(name string) {
	log.WithFields(log.Fields{"name": name, "labels": config.Labels, "command": config.Command}).Info("config server")

	for key, value := range config.Arguments {
		log.WithFields(log.Fields{"name": name, "argument": key, "value": value}).Info("config server argument")
	}
}

type Node struct {
	IP     string `yaml:"ip"`
	Index  uint   `yaml:"index"`
	Labels Labels `yaml:"labels"`
}

func NewNode(ip string, index uint, labels []string) *Node {
	return &Node{IP: ip, Index: index, Labels: labels}
}

func (node *Node) IsController() bool {
	for _, label := range node.Labels {
		if label == utils.NODE_CONTROLLER {
			return true
		}
	}

	return false
}

type Nodes map[string]*Node
type Commands map[string]*Command

type Config struct {
	Version         string                     `yaml:"version"`
	APIServerPort   uint16                     `yaml:"apiserver-port"`
	DeploymentFiles map[string]*DeploymentFile `yaml:"deployment-files"`
	Nodes           Nodes                      `yaml:"nodes,omitempty"`
	Commands        Commands                   `yaml:"commands,omitempty"`
	Servers         map[string]*ServerConfig   `yaml:"servers,omitempty"`
}

type InternalConfig struct {
	BaseDirectory string
	Name          string
	Node          *Node
	Config        *Config
}

func (config *InternalConfig) GetTemplateDeploymentFilename(name string) string {
	return fmt.Sprintf(`{{deployment_file "%s"}}`, name)
}

func (config *InternalConfig) GetFullTargetDeploymentFilename(name string) string {
	var result *DeploymentFile
	var ok bool

	if result, ok = config.Config.DeploymentFiles[name]; !ok {
		log.WithFields(log.Fields{"name": name}).Fatal("missing deployment file")
	}

	resultFilename, error := config.ApplyTemplate("deployment-file", result.File)
	if error != nil {
		log.WithFields(log.Fields{"name": name, "error": error}).Fatal("deployment file expansion")
	}

	return path.Join("/", resultFilename)
}

func (config *InternalConfig) GetFullDeploymentFilename(name string) string {
	var result *DeploymentFile
	var ok bool

	if result, ok = config.Config.DeploymentFiles[name]; !ok {
		log.WithFields(log.Fields{"name": name}).Fatal("missing deployment file")
	}

	resultFilename, error := config.ApplyTemplate("deployment-file", result.File)

	if error != nil {
		log.WithFields(log.Fields{"name": name, "error": error}).Fatal("deployment file expansion")
	}

	resultFilename = path.Join(config.BaseDirectory, resultFilename)

	directory := path.Dir(resultFilename)

	if error := utils.CreateDirectoryIfMissing(directory); error != nil {
		log.WithFields(log.Fields{"directory": directory, "error": error}).Fatal("could not create directory")
	}

	return resultFilename
}

func (config *InternalConfig) SetNode(nodeName string, node *Node) {
	config.Name = nodeName
	config.Node = node
}

func DefaultInternalConfig(baseDirectory string) *InternalConfig {
	config := &InternalConfig{}
	config.BaseDirectory = baseDirectory
	config.Config = &Config{Version: utils.CONFIG_VERSION}
	config.Config.DeploymentFiles = map[string]*DeploymentFile{}
	config.Config.Nodes = Nodes{}
	config.Config.Commands = Commands{}
	config.Config.Servers = map[string]*ServerConfig{}

	// Default Settings
	config.Config.APIServerPort = 6443

	// Config
	config.addDeploymentFile(utils.CONFIG_FILENAME, Labels{utils.NODE_CONTROLLER, utils.NODE_WORKER}, utils.GetFullConfigDirectory())

	// Binaries
	config.addDeploymentFile(utils.K8S_TEW_BINARY, Labels{utils.NODE_CONTROLLER, utils.NODE_WORKER}, utils.GetFullBinariesDirectory())

	// CNI Binaries
	config.addDeploymentFile(utils.BRIDGE_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCNIBinariesDirectory())
	config.addDeploymentFile(utils.FLANNEL_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCNIBinariesDirectory())
	config.addDeploymentFile(utils.LOOPBACK_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCNIBinariesDirectory())
	config.addDeploymentFile(utils.HOST_LOCAL_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCNIBinariesDirectory())

	// ContainerD Binaries
	config.addDeploymentFile(utils.CONTAINERD_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCRIBinariesDirectory())
	config.addDeploymentFile(utils.CONTAINERD_SHIM_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCRIBinariesDirectory())
	config.addDeploymentFile(utils.CRI_CONTAINERD_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCRIBinariesDirectory())
	config.addDeploymentFile(utils.CRICTL_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCRIBinariesDirectory())
	config.addDeploymentFile(utils.CTR_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCRIBinariesDirectory())
	config.addDeploymentFile(utils.RUNC_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullCRIBinariesDirectory())

	// Etcd Binaries
	config.addDeploymentFile(utils.ETCD_BINARY, Labels{utils.NODE_CONTROLLER}, utils.GetFullETCDBinariesDirectory())
	config.addDeploymentFile(utils.ETCDCTL_BINARY, Labels{utils.NODE_CONTROLLER}, utils.GetFullETCDBinariesDirectory())
	config.addDeploymentFile(utils.FLANNELD_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullETCDBinariesDirectory())

	// K8S Binaries
	config.addDeploymentFile(utils.KUBECTL_BINARY, Labels{utils.NODE_CONTROLLER}, utils.GetFullK8SBinariesDirectory())
	config.addDeploymentFile(utils.KUBE_APISERVER_BINARY, Labels{utils.NODE_CONTROLLER}, utils.GetFullK8SBinariesDirectory())
	config.addDeploymentFile(utils.KUBE_CONTROLLER_MANAGER_BINARY, Labels{utils.NODE_CONTROLLER}, utils.GetFullK8SBinariesDirectory())
	config.addDeploymentFile(utils.KUBELET_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullK8SBinariesDirectory())
	config.addDeploymentFile(utils.KUBE_PROXY_BINARY, Labels{utils.NODE_WORKER}, utils.GetFullK8SBinariesDirectory())
	config.addDeploymentFile(utils.KUBE_SCHEDULER_BINARY, Labels{utils.NODE_CONTROLLER}, utils.GetFullK8SBinariesDirectory())

	// Certificates
	config.addDeploymentFile(utils.CA_PEM, Labels{utils.NODE_CONTROLLER, utils.NODE_WORKER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.CA_KEY_PEM, Labels{utils.NODE_CONTROLLER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.KUBERNETES_PEM, Labels{utils.NODE_CONTROLLER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.KUBERNETES_KEY_PEM, Labels{utils.NODE_CONTROLLER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.ADMIN_PEM, Labels{}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.ADMIN_KEY_PEM, Labels{}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.PROXY_PEM, Labels{utils.NODE_WORKER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.PROXY_KEY_PEM, Labels{utils.NODE_WORKER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.KUBELET_PEM, Labels{utils.NODE_WORKER}, utils.GetFullCertificatesConfigDirectory())
	config.addDeploymentFile(utils.KUBELET_KEY_PEM, Labels{utils.NODE_WORKER}, utils.GetFullCertificatesConfigDirectory())

	// Kubeconfig
	config.addDeploymentFile(utils.ADMIN_KUBECONFIG, Labels{}, utils.GetFullKubeConfigDirectory())
	config.addDeploymentFile(utils.PROXY_KUBECONFIG, Labels{utils.NODE_WORKER}, utils.GetFullKubeConfigDirectory())
	config.addDeploymentFile(utils.KUBELET_KUBECONFIG, Labels{utils.NODE_WORKER}, utils.GetFullKubeConfigDirectory())

	// Security
	config.addDeploymentFile(utils.ENCRYPTION_CONFIG, Labels{utils.NODE_CONTROLLER}, utils.GetFullSecurityConfigDirectory())

	// CNI
	config.addDeploymentFile(utils.NET_CONFIG, Labels{utils.NODE_WORKER}, utils.GetFullCNIConfigDirectory())
	config.addDeploymentFile(utils.CNI_CONFIG, Labels{utils.NODE_WORKER}, utils.GetFullCNIConfigDirectory())

	// Service
	config.addDeploymentFile(utils.SERVICE_CONFIG, Labels{utils.NODE_CONTROLLER, utils.NODE_WORKER}, utils.GetFullServiceDirectory())

	// K8S Config
	config.addDeploymentFile(utils.K8S_KUBELET_CONFIG, Labels{utils.NODE_CONTROLLER}, utils.GetFullK8SConfigDirectory())
	config.addDeploymentFile(utils.K8S_ADMIN_USER_CONFIG, Labels{utils.NODE_CONTROLLER}, utils.GetFullK8SConfigDirectory())

	// Dependencies
	config.addCommand("load-overlay", Labels{utils.NODE_CONTROLLER, utils.NODE_WORKER}, "modprobe overlay")
	config.addCommand("flanneld-configuration", Labels{utils.NODE_CONTROLLER}, fmt.Sprintf("%s --ca-file=%s --cert-file=%s --key-file=%s set /coreos.com/network/config '{ \"Network\": \"%s.0.0/16\" }'", config.GetTemplateDeploymentFilename(utils.ETCDCTL_BINARY), config.GetTemplateDeploymentFilename(utils.CA_PEM), config.GetTemplateDeploymentFilename(utils.KUBERNETES_PEM), config.GetTemplateDeploymentFilename(utils.KUBERNETES_KEY_PEM), utils.CIDR_PREFIX))
	config.addCommand("k8s-kubelet-configuration", Labels{utils.NODE_CONTROLLER}, fmt.Sprintf("%s --server=127.0.0.1:8080 apply -f %s", config.GetTemplateDeploymentFilename(utils.KUBECTL_BINARY), config.GetTemplateDeploymentFilename(utils.K8S_KUBELET_CONFIG)))
	config.addCommand("k8s-admin-user-configuration", Labels{utils.NODE_CONTROLLER}, fmt.Sprintf("%s --server=127.0.0.1:8080 apply -f %s", config.GetTemplateDeploymentFilename(utils.KUBECTL_BINARY), config.GetTemplateDeploymentFilename(utils.K8S_ADMIN_USER_CONFIG)))
	config.addCommand("k8s-kube-dns", Labels{utils.NODE_CONTROLLER}, fmt.Sprintf("%s --server=127.0.0.1:8080 apply -f https://storage.googleapis.com/kubernetes-the-hard-way/kube-dns.yaml", config.GetTemplateDeploymentFilename(utils.KUBECTL_BINARY)))
	config.addCommand("k8s-kubernetes-dashboard", Labels{utils.NODE_CONTROLLER}, fmt.Sprintf("%s --server=127.0.0.1:8080 apply -f https://raw.githubusercontent.com/kubernetes/dashboard/master/src/deploy/recommended/kubernetes-dashboard.yaml", config.GetTemplateDeploymentFilename(utils.KUBECTL_BINARY)))

	// Servers
	config.addServer("etcd", Labels{utils.NODE_CONTROLLER}, config.GetTemplateDeploymentFilename(utils.ETCD_BINARY), map[string]string{
		"name":                        "{{.Name}}",
		"cert-file":                   config.GetTemplateDeploymentFilename(utils.KUBERNETES_PEM),
		"key-file":                    config.GetTemplateDeploymentFilename(utils.KUBERNETES_KEY_PEM),
		"peer-cert-file":              config.GetTemplateDeploymentFilename(utils.KUBERNETES_PEM),
		"peer-key-file":               config.GetTemplateDeploymentFilename(utils.KUBERNETES_KEY_PEM),
		"trusted-ca-file":             config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"peer-trusted-ca-file":        config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"peer-client-cert-auth":       "",
		"client-cert-auth":            "",
		"initial-advertise-peer-urls": "https://{{.Node.IP}}:2380",
		"listen-peer-urls":            "https://{{.Node.IP}}:2380",
		"listen-client-urls":          "https://{{.Node.IP}}:2379,http://127.0.0.1:2379",
		"advertise-client-urls":       "https://{{.Node.IP}}:2379",
		"initial-cluster-token":       "etcd-cluster",
		"initial-cluster":             "{{etcd_cluster}}",
		"initial-cluster-state":       "new",
		"data-dir":                    path.Join("{{.BaseDirectory}}", utils.GetFullETCDDataDirectory()),
	})

	config.addServer("kube-apiserver", Labels{utils.NODE_CONTROLLER}, config.GetTemplateDeploymentFilename(utils.KUBE_APISERVER_BINARY), map[string]string{
		"allow-privileged":                        "true",
		"admission-control":                       "Initializers,NamespaceLifecycle,NodeRestriction,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota",
		"advertise-address":                       "{{.Node.IP}}",
		"apiserver-count":                         "{{controllers_count}}",
		"audit-log-maxage":                        "30",
		"audit-log-maxbackup":                     "3",
		"audit-log-maxsize":                       "100",
		"audit-log-path":                          path.Join("{{.BaseDirectory}}", utils.GetFullLoggingDirectory(), utils.AUDIT_LOG),
		"authorization-mode":                      "Node,RBAC",
		"bind-address":                            "{{.Node.IP}}",
		"secure-port":                             "{{.Config.APIServerPort}}",
		"client-ca-file":                          config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"enable-swagger-ui":                       "true",
		"etcd-cafile":                             config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"etcd-certfile":                           config.GetTemplateDeploymentFilename(utils.KUBERNETES_PEM),
		"etcd-keyfile":                            config.GetTemplateDeploymentFilename(utils.KUBERNETES_KEY_PEM),
		"etcd-servers":                            "{{etcd_servers}}",
		"event-ttl":                               "1h",
		"experimental-encryption-provider-config": config.GetTemplateDeploymentFilename(utils.ENCRYPTION_CONFIG),
		"insecure-bind-address":                   "127.0.0.1",
		"kubelet-certificate-authority":           config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"kubelet-client-certificate":              config.GetTemplateDeploymentFilename(utils.KUBERNETES_PEM),
		"kubelet-client-key":                      config.GetTemplateDeploymentFilename(utils.KUBERNETES_KEY_PEM),
		"kubelet-https":                           "true",
		"runtime-config":                          "api/all",
		"service-account-key-file":                config.GetTemplateDeploymentFilename(utils.CA_KEY_PEM),
		"service-cluster-ip-range":                "10.32.0.0/24",
		"service-node-port-range":                 "30000-32767",
		"tls-ca-file":                             config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"tls-cert-file":                           config.GetTemplateDeploymentFilename(utils.KUBERNETES_PEM),
		"tls-private-key-file":                    config.GetTemplateDeploymentFilename(utils.KUBERNETES_KEY_PEM),
		"v": "0",
	})

	config.addServer("kube-controller-manager", Labels{utils.NODE_CONTROLLER}, config.GetTemplateDeploymentFilename(utils.KUBE_CONTROLLER_MANAGER_BINARY), map[string]string{
		"address":                          "0.0.0.0",
		"cluster-cidr":                     fmt.Sprintf("%s.0.0/16", utils.CIDR_PREFIX),
		"cluster-name":                     "kubernetes",
		"cluster-signing-cert-file":        config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"cluster-signing-key-file":         config.GetTemplateDeploymentFilename(utils.CA_KEY_PEM),
		"leader-elect":                     "true",
		"master":                           "http://127.0.0.1:8080",
		"root-ca-file":                     config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"service-account-private-key-file": config.GetTemplateDeploymentFilename(utils.CA_KEY_PEM),
		"service-cluster-ip-range":         "10.32.0.0/24",
		"v": "0",
	})

	config.addServer("kube-scheduler", Labels{utils.NODE_CONTROLLER}, config.GetTemplateDeploymentFilename(utils.KUBE_SCHEDULER_BINARY), map[string]string{
		"leader-elect": "true",
		"master":       "http://127.0.0.1:8080",
		"v":            "0",
	})

	config.addServer("kubelet", Labels{utils.NODE_WORKER}, config.GetTemplateDeploymentFilename(utils.KUBELET_BINARY), map[string]string{
		"allow-privileged":             "true",
		"anonymous-auth":               "false",
		"authorization-mode":           "Webhook",
		"client-ca-file":               config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"cluster-dns":                  "10.32.0.10",
		"cluster-domain":               "cluster.local",
		"container-runtime":            "remote",
		"container-runtime-endpoint":   "unix:///var/run/cri-containerd.sock",
		"image-pull-progress-deadline": "2m",
		"kubeconfig":                   config.GetTemplateDeploymentFilename(utils.KUBELET_KUBECONFIG),
		"network-plugin":               "cni",
		"pod-cidr":                     fmt.Sprintf("%s.{{.Node.Index}}.0/24", utils.CIDR_PREFIX),
		"register-node":                "true",
		"require-kubeconfig":           "",
		"runtime-request-timeout":      "15m",
		"tls-cert-file":                config.GetTemplateDeploymentFilename(utils.KUBELET_PEM),
		"tls-private-key-file":         config.GetTemplateDeploymentFilename(utils.KUBELET_KEY_PEM),
		"v": "0",
	})

	config.addServer("kube-proxy", Labels{utils.NODE_WORKER}, config.GetTemplateDeploymentFilename(utils.KUBE_PROXY_BINARY), map[string]string{
		"cluster-cidr": fmt.Sprintf("%s.0.0/16", utils.CIDR_PREFIX),
		"kubeconfig":   config.GetTemplateDeploymentFilename(utils.PROXY_KUBECONFIG),
		"proxy-mode":   "iptables",
		"v":            "0",
	})

	config.addServer("cri-containerd", Labels{utils.NODE_WORKER}, config.GetTemplateDeploymentFilename(utils.CRI_CONTAINERD_BINARY), map[string]string{
		"network-conf-dir": path.Join("{{.BaseDirectory}}", utils.GetFullCNIConfigDirectory()),
		"network-bin-dir":  path.Join("{{.BaseDirectory}}", utils.GetFullCNIBinariesDirectory()),
	})

	config.addServer("containerd", Labels{utils.NODE_WORKER}, config.GetTemplateDeploymentFilename(utils.CONTAINERD_BINARY), map[string]string{})

	config.addServer("flanneld", Labels{utils.NODE_WORKER}, config.GetTemplateDeploymentFilename(utils.FLANNELD_BINARY), map[string]string{
		"etcd-endpoints": "{{etcd_servers}}",
		"etcd-cafile":    config.GetTemplateDeploymentFilename(utils.CA_PEM),
		"etcd-certfile":  config.GetTemplateDeploymentFilename(utils.KUBELET_PEM),
		"etcd-keyfile":   config.GetTemplateDeploymentFilename(utils.KUBELET_KEY_PEM),
		"v":              "0",
	})

	return config
}

func (config *InternalConfig) GetForwarderAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", config.Config.APIServerPort)
}

func (config *InternalConfig) addServer(name string, labels []string, command string, arguments map[string]string) {
	config.Config.Servers[name] = &ServerConfig{Labels: labels, Command: command, Arguments: arguments}
}

func (config *InternalConfig) addCommand(name string, labels Labels, command string) {
	config.Config.Commands[name] = NewCommand(labels, command)
}

func (config *InternalConfig) addDeploymentFile(name string, labels Labels, _path string) {
	config.Config.DeploymentFiles[name] = NewDeploymentFile(labels, path.Join(_path, name))
}

func (config *InternalConfig) Dump() {
	log.WithFields(log.Fields{"base-directory": config.BaseDirectory}).Info("config")
	log.WithFields(log.Fields{"name": config.Name}).Info("config")

	if config.Node != nil {
		log.WithFields(log.Fields{"ip": config.Node.IP}).Info("config")
		log.WithFields(log.Fields{"labels": config.Node.Labels}).Info("config")
		log.WithFields(log.Fields{"index": config.Node.Index}).Info("config")
	}

	for name, deploymentFile := range config.Config.DeploymentFiles {
		log.WithFields(log.Fields{"name": name, "file": deploymentFile.File, "labels": deploymentFile.Labels}).Info("config deployment file")
	}

	for name, node := range config.Config.Nodes {
		log.WithFields(log.Fields{"name": name, "index": node.Index, "labels": node.Labels, "ip": node.IP}).Info("config node")
	}

	for name, command := range config.Config.Commands {
		log.WithFields(log.Fields{"name": name, "command": command.Command, "labels": command.Labels}).Info("config command")
	}

	for name, serverConfig := range config.Config.Servers {
		serverConfig.Dump(name)
	}
}

func (config *InternalConfig) getConfigDirectory() string {
	return path.Join(config.BaseDirectory, utils.GetFullConfigDirectory())
}

func (config *InternalConfig) getConfigFilename() string {
	return path.Join(config.getConfigDirectory(), utils.CONFIG_FILENAME)
}

func (config *InternalConfig) Save() error {
	if error := utils.CreateDirectoryIfMissing(config.getConfigDirectory()); error != nil {
		return error
	}

	yamlOutput, error := yaml.Marshal(config.Config)

	if error != nil {
		return error
	}

	filename := config.getConfigFilename()

	return ioutil.WriteFile(filename, yamlOutput, 0644)
}

func (config *InternalConfig) Load() error {
	var error error

	filename := config.getConfigFilename()

	// Check if config file exists
	if _, error := os.Stat(filename); os.IsNotExist(error) {
		return errors.New(fmt.Sprintf("config '%s' not found", filename))
	}

	yamlContent, error := ioutil.ReadFile(filename)

	if error != nil {
		return error
	}

	if error := yaml.Unmarshal(yamlContent, config.Config); error != nil {
		return error
	}

	if len(config.Name) == 0 {
		config.Name, error = os.Hostname()

		if error != nil {
			return error
		}
	}

	if config.Node == nil {
		for name, node := range config.Config.Nodes {
			if name != config.Name {
				continue
			}

			config.Node = node

			break
		}
	}

	return nil
}

func (config *InternalConfig) RemoveNode(name string) error {
	if _, ok := config.Config.Nodes[name]; !ok {
		return errors.New("node not found")
	}

	delete(config.Config.Nodes, name)

	return nil
}

func (config *InternalConfig) AddNode(name string, ip string, index uint, labels []string) (*Node, error) {
	name = strings.Trim(name, " \n")

	if len(name) == 0 {
		return nil, errors.New("empty node name")
	}

	if net.ParseIP(ip) == nil {
		return nil, errors.New("invalid or wrong ip format")
	}

	config.Config.Nodes[name] = NewNode(ip, index, labels)

	return config.Config.Nodes[name], nil
}

func (config *InternalConfig) ApplyTemplate(label string, value string) (string, error) {
	var functions = template.FuncMap{
		"controllers_count": func() string {
			count := 0
			for _, node := range config.Config.Nodes {
				if node.IsController() {
					count += 1
				}
			}

			return fmt.Sprintf("%d", count)
		},
		"etcd_servers": func() string {
			result := ""

			for _, node := range config.Config.Nodes {
				if !node.IsController() {
					continue
				}

				if len(result) > 0 {
					result += ","
				}

				result += fmt.Sprintf("https://%s:2379", node.IP)
			}

			return result
		},
		"etcd_cluster": func() string {
			result := ""

			for name, node := range config.Config.Nodes {
				if !node.IsController() {
					continue
				}

				if len(result) > 0 {
					result += ","
				}

				result += fmt.Sprintf("%s=https://%s:2380", name, node.IP)
			}

			return result
		},
		"deployment_file": func(name string) string {
			return config.GetFullDeploymentFilename(name)
		},
	}

	var newValue bytes.Buffer

	argumentTemplate, error := template.New(fmt.Sprintf(label)).Funcs(functions).Parse(value)

	if error != nil {
		return "", error
	}

	if error = argumentTemplate.Execute(&newValue, config); error != nil {
		return "", error
	}

	return newValue.String(), nil
}

func CompareLabels(source, destination Labels) bool {
	return source != nil && destination != nil && source.HasLabels(destination)
}
